package gorm

import (
	"context"
	"database/sql"
	"errors"
	"github.com/gogf/gf/v2/container/gmap"
	"github.com/gogf/gf/v2/container/gset"
	"github.com/powershopy/library/logging"
	"github.com/powershopy/library/utils"
	"gorm.io/gorm"
	"math/rand"
	"sync"
	"time"
)

type pwPolicy struct {
	readFailNodes     *gset.IntSet
	readPendingNodes  *gmap.IntIntMap
	writeFailNodes    *gset.IntSet
	writePendingNodes *gmap.IntIntMap
	readNodes         []gorm.ConnPool
	writeNodes        []gorm.ConnPool
	duration          time.Duration
	Name              string
	readLen           int
	writeLen          int
	loadNode          bool
}

func NewPolicy(read []gorm.Dialector, write []gorm.Dialector, t time.Duration, name string) (*pwPolicy, error) {
	readLen := len(read)
	writeLen := len(write)
	if readLen == writeLen {
		//Dialector推不出connpool 与sql.DB 反过来也推不出
		return nil, errors.New("目前不支持读取节点数与写入节点数一致")
	}
	p := &pwPolicy{
		readFailNodes:     gset.NewIntSet(true),
		readPendingNodes:  gmap.NewIntIntMap(true),
		writeFailNodes:    gset.NewIntSet(true),
		writePendingNodes: gmap.NewIntIntMap(true),
		readNodes:         []gorm.ConnPool{},
		writeNodes:        []gorm.ConnPool{},
		duration:          t,
		Name:              name,
		readLen:           readLen,
		writeLen:          writeLen,
	}
	go p.checkNodes()
	return p, nil
}

func (p *pwPolicy) Resolve(connPools []gorm.ConnPool) gorm.ConnPool {
	if !p.loadNode {
		if len(connPools) == p.readLen {
			p.readNodes = connPools
		}
		if len(connPools) == p.writeLen {
			p.writeNodes = connPools
		}
		if p.writeLen > 0 && p.readLen > 0 {
			p.loadNode = true
		}
	}
	index := 0
	if len(connPools) == p.readLen { //读
		if p.readFailNodes.Size() >= len(connPools) { //节点全部不可用
			index = rand.Intn(len(connPools))
		} else {
			index = utils.RandInt(len(connPools), p.readFailNodes.Slice()...)
		}
	} else { //写
		if p.writeFailNodes.Size() >= len(connPools) {
			index = rand.Intn(len(connPools))
		} else {
			index = utils.RandInt(len(connPools), p.writeFailNodes.Slice()...)
		}
	}

	//logging.WithField("index", index).Debug(context.Background(), "resovle")
	return connPools[index]
}

func (p *pwPolicy) checkNodes() {
	for {
		time.Sleep(p.duration) //每5s进行一次check
		if len(p.readNodes) <= 0 && len(p.writeNodes) <= 0 {
			continue
		}
		wg := sync.WaitGroup{}
		for index, pool := range p.readNodes {
			db, ok := pool.(*sql.DB)
			if ok {
				logging.WithFields(map[string]interface{}{
					"index": index,
					"name":  p.Name,
					"type":  "read",
				}).Debug(context.Background(), "ping...")
				wg.Add(1)
				go func(index int, db *sql.DB) {
					defer wg.Done()
					ctx, _ := context.WithTimeout(context.Background(), time.Second)
					err := db.PingContext(ctx)
					if err != nil {
						logging.WithFields(map[string]interface{}{
							"name":  p.Name,
							"index": index,
							"type":  "read",
						}).Error(context.Background(), "db node down")
						p.readFailNodes.Add(index)
					} else {
						if p.readPendingNodes.Contains(index) {
							if p.readPendingNodes.Get(index) >= 2 { //连续3次成功
								p.readPendingNodes.Remove(index)
								p.readFailNodes.Remove(index) //移出失败节点列表
							} else {
								p.readPendingNodes.Set(index, p.readPendingNodes.Get(index)+1) //加权
							}
						} else if p.readFailNodes.Contains(index) {
							p.readPendingNodes.Set(index, 1) //失败后第一次成功
						}
					}
				}(index, db)
			}
		}
		for index, pool := range p.writeNodes { //写
			db, ok := pool.(*sql.DB)
			if ok {
				logging.WithFields(map[string]interface{}{
					"index": index,
					"name":  p.Name,
					"type":  "write",
				}).Debug(context.Background(), "ping...")
				wg.Add(1)
				go func(index int, db *sql.DB) {
					defer wg.Done()
					ctx, _ := context.WithTimeout(context.Background(), time.Second)
					err := db.PingContext(ctx)
					if err != nil {
						logging.WithFields(map[string]interface{}{
							"name":  p.Name,
							"index": index,
							"type":  "write",
						}).Error(context.Background(), "db node down")
						p.writeFailNodes.Add(index)
					} else {
						if p.writePendingNodes.Contains(index) {
							if p.writePendingNodes.Get(index) >= 2 { //连续3次成功
								p.writePendingNodes.Remove(index)
								p.writeFailNodes.Remove(index) //移出失败节点列表
							} else {
								p.writePendingNodes.Set(index, p.writePendingNodes.Get(index)+1) //加权
							}
						} else if p.writeFailNodes.Contains(index) {
							p.writePendingNodes.Set(index, 1) //失败后第一次成功
						}
					}
				}(index, db)
			}
		}
		wg.Wait()
	}
}
