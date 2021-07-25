package main

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/progrium/macdriver/cocoa"
	"github.com/progrium/macdriver/core"
	"github.com/progrium/macdriver/objc"
)

type NSUserNotification struct {
	objc.Object
}

var NSUserNotification_ = objc.Get("NSUserNotification")

type NSUserNotificationCenter struct {
	objc.Object
}

var NSUserNotificationCenter_ = objc.Get("NSUserNotificationCenter")

func main() {
	runtime.LockOSThread()

	app := cocoa.NSApp_WithDidLaunch(func(n objc.Object) {
		obj := cocoa.NSStatusBar_System().StatusItemWithLength(cocoa.NSVariableStatusItemLength)
		obj.Retain()
		obj.Button().SetTitle("🚀️ Ready")
		/*
			nextClicked := make(chan bool)
			go func() {
				state := -1
				timer := 1500
				countdown := false
				for {
					select {
					case <-time.After(1 * time.Second):
						if timer > 0 && countdown {
							timer = timer - 1
						}
						if timer <= 0 && state%2 == 1 {
							state = (state + 1) % 4
						}
					case <-nextClicked:
						state = (state + 1) % 4
						timer = map[int]int{
							0: 10,
							1: 10,
							2: 0,
							3: 3,
						}[state]
						if state%2 == 1 {
							countdown = true
						} else {
							countdown = false
						}
					}
					labels := map[int]string{
						0: "▶️ Ready %02d:%02d",
						1: "✴️ Working %02d:%02d",
						2: "✅ Finished %02d:%02d",
						3: "⏸️ Break %02d:%02d",
					}
					// updates to the ui should happen on the main thread to avoid strange bugs
					core.Dispatch(func() {
						obj.Button().SetTitle(fmt.Sprintf(labels[state], timer/60, timer%60))

					})
				}
			}()
			nextClicked <- true
		*/

		var cancelTimerFunc func()
		var ctx context.Context
		timerStarted := false

		itemStartStop := cocoa.NSMenuItem_New()
		itemStartStop.SetTitle("Start")
		itemStartStop.SetAction(objc.Sel("nextClicked:"))
		cocoa.DefaultDelegateClass.AddMethod("nextClicked:", func(_ objc.Object) {

			log.Print("timerStarted: ", timerStarted)

			if !timerStarted {
				timerDone := make(chan int)
				ctx, cancelTimerFunc = context.WithCancel(context.Background())
				go StartTimer(ctx, 10*time.Second, "prefix ", &obj, timerDone)

				timerStarted = true

				go func() {

					defer func() {
						itemStartStop.SetTitle("Start")
						obj.Button().SetTitle("🚀️ Ready")
						timerStarted = false
					}()

					select {
					case <-timerDone:
						notification := NSUserNotification{NSUserNotification_.Alloc().Init()}
						notification.Set("title:", core.String("Hello, sdfgsfg!"))
						notification.Set("informativeText:", core.String("More text"))

						center := NSUserNotificationCenter{NSUserNotificationCenter_.Send("defaultUserNotificationCenter")}
						center.Send("deliverNotification:", notification)
						notification.Release()

						log.Print("timerdone: ", len(timerDone))

					case <-ctx.Done():
						log.Print("ctx")
					}
					log.Print("end")
				}()

				itemStartStop.SetTitle("Stop")
			} else {
				cancelTimerFunc()
				itemStartStop.SetTitle("Start")
			}
		})

		itemQuit := cocoa.NSMenuItem_New()
		itemQuit.SetTitle("Quit")
		itemQuit.SetAction(objc.Sel("terminate:"))

		menu := cocoa.NSMenu_New()
		menu.AddItem(itemStartStop)
		menu.AddItem(itemQuit)
		obj.SetMenu(menu)

	})

	nsbundle := cocoa.NSBundle_Main().Class()

	nsbundle.AddMethod("__bundleIdentifier", func(_ objc.Object) objc.Object {
		return core.String("me.bn4t.macdriver-app")
	})
	nsbundle.Swizzle("bundleIdentifier", "__bundleIdentifier")

	app.SetActivationPolicy(cocoa.NSApplicationActivationPolicyRegular)
	app.ActivateIgnoringOtherApps(true)
	app.Run()
}

func StartTimer(ctx context.Context, duration time.Duration, prefix string, statusItem *cocoa.NSStatusItem, done chan int) {
	ticker := time.NewTicker(1 * time.Second)
	remainingTime := int(duration.Seconds())

	// updates to the ui should happen on the main thread to avoid strange bugs
	core.Dispatch(func() {
		statusItem.Button().SetTitle(fmt.Sprintf(prefix+"%02d:%02d", remainingTime/60, remainingTime%60))
	})

	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return

		case <-ticker.C:
			remainingTime--

			uiUpdate := make(chan int)
			// updates to the ui should happen on the main thread to avoid strange bugs
			core.Dispatch(func() {
				statusItem.Button().SetTitle(fmt.Sprintf(prefix+"%02d:%02d", remainingTime/60, remainingTime%60))
				uiUpdate <- 1
			})

			<-uiUpdate

			if remainingTime <= 0 {
				done <- 1
				ticker.Stop()
				return
			}
		}
	}

}
