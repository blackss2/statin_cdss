// +build windows

package util

import (
	"github.com/rjeczalik/notify"
)

func init() {
	WatcherNotifies = []notify.Event{notify.All, notify.FileNotifyChangeLastWrite}
}
