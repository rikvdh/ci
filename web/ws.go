package web

import (
	"encoding/json"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/go-iris2/iris2"
	"github.com/go-iris2/iris2/adaptors/websocket"
	"github.com/rikvdh/ci/lib/builder"
	"github.com/rikvdh/ci/lib/indexer"
	"github.com/rikvdh/ci/models"
)

func getBuildList() []byte {
	var msg struct {
		Running []models.Job `json:"running"`
		Queue   []models.Job `json:"queue"`
	}

	models.Handle().Where("status = ?", models.StatusBusy).Order("start DESC").Find(&msg.Running)
	models.Handle().Where("status = ?", models.StatusNew).Order("start DESC").Find(&msg.Queue)

	data, _ := json.Marshal(msg)
	return data
}

func startWs(app *iris2.Framework) {
	ws := websocket.New(websocket.Config{
		Endpoint:         "/ws",
		ClientSourcePath: "/ws.js",
	})

	app.Adapt(ws)

	var mu sync.RWMutex

	conns := make(map[string]websocket.Connection)

	go func() {
		ch := builder.GetEventChannel()
		for {
			<-*ch

			data := getBuildList()
			logrus.Infof("Emit new build-list via websocket to %d connections", len(conns))
			mu.RLock()
			for _, con := range conns {
				con.EmitMessage(data)
			}
			mu.RUnlock()
		}
	}()

	ws.OnConnection(func(c websocket.Connection) {
		ctx := c.Context()
		if ctx.Session().GetString("authenticated") != "true" {
			c.Disconnect()
			return
		}
		mu.Lock()
		conns[c.ID()] = c
		mu.Unlock()

		c.EmitMessage(getBuildList())

		c.OnMessage(func(d []byte) {
			var req struct {
				Action string
				ID     int `json:",omitempty"`
			}
			json.Unmarshal(d, &req)
			if req.Action == "build" {
				item := models.Branch{}
				models.Handle().Preload("Build").Where("id = ?", req.ID).First(&item)
				if item.ID > 0 {
					indexer.ScheduleJob(item.Build.ID, item.ID, item.LastReference)
				} else {
					logrus.Warnf("error, branch not found: %v", req)
				}
			}
		})

		c.OnDisconnect(func() {
			mu.Lock()
			delete(conns, c.ID())
			mu.Unlock()
		})
	})
}
