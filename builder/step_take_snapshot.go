package profitbricks

import (
	"errors"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/profitbricks/profitbricks-sdk-go"
	"strings"
	"time"
)

type stepTakeSnapshot struct{}

func (s *stepTakeSnapshot) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(*Config)

	ui.Say("Creating Profitbricks snapshot...")

	profitbricks.SetAuth(c.PBUsername, c.PBPassword)

	dcId := state.Get("datacenter_id").(string)
	volumeId := state.Get("volume_id").(string)

	snapshot := profitbricks.CreateSnapshot(dcId, volumeId, c.ServerName+"Snapshot")

	state.Put("snapshotname", c.ServerName+"Snapshot")

	err := s.checkForErrors(snapshot)
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.waitTillProvisioned(strings.Join(snapshot.Headers["Location"], ""), *c)

	return multistep.ActionContinue
}

func (s *stepTakeSnapshot) Cleanup(state multistep.StateBag) {
}

func (d *stepTakeSnapshot) checkForErrors(instance profitbricks.Resp) error {
	if instance.StatusCode > 299 {
		return errors.New(fmt.Sprintf("Error occured %s", string(instance.Body)))
	}
	return nil
}

func (d *stepTakeSnapshot) waitTillProvisioned(path string, config Config) {
	d.setPB(config.PBUsername, config.PBPassword, config.PBUrl)
	for i := 0; i < waitCount; i++ {
		request := profitbricks.GetRequestStatus(path)
		if request.MetaData["status"] == "DONE" {
			break
		}
		time.Sleep(10 * time.Second)
		i++
	}
}

func (d *stepTakeSnapshot) setPB(username string, password string, url string) {
	profitbricks.SetAuth(username, password)
	profitbricks.SetEndpoint(url)
}
