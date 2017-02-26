package web

import (
	"encoding/json"
	"fmt"
	"github.com/go-iris2/iris2"
	"github.com/go-iris2/iris2/adaptors/websocket"
	"github.com/rikvdh/ci/lib/builder"
	"github.com/rikvdh/ci/lib/indexer"
	"github.com/rikvdh/ci/models"
	"sync"
)

func startWs(app *iris2.Framework) {
	ws := websocket.New(websocket.Config{
		Endpoint:         "/ws",
		ClientSourcePath: "/ws.js",
	})

	app.Adapt(ws)

	var conns map[string]websocket.Connection
	var mu sync.Mutex

	conns = make(map[string]websocket.Connection)

	go func() {
		ch := builder.GetEventChannel()
		for {
			var running []models.Job

			<-ch
			models.Handle().Where("status = ?", models.StatusBusy).Order("start DESC").Find(&running)
			data, _ := json.Marshal(running)

			mu.Lock()
			for _, con := range conns {
				con.EmitMessage(data)
			}
			mu.Unlock()
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

		var running []models.Job
		models.Handle().Where("status = ?", models.StatusBusy).Order("start DESC").Find(&running)
		data, _ := json.Marshal(running)
		c.EmitMessage(data)

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
					fmt.Printf("error, branch not found: %v\n", req)
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
