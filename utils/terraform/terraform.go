package terraform

import (
	"io/ioutil"
	"log"

	ps "github.com/bhendo/go-powershell"
	"github.com/bhendo/go-powershell/backend"
)

// tfInit will initialize the new state.
func Init(statepath string) {
	back := &backend.Local{}
	shell, err := ps.New(back)

	check(err)

	defer shell.Exit()

	_, _, err = shell.Execute("tf init -reconfigure " + statepath)
	if err != nil {
		log.Printf(err.Error())
	}

}

// tfCreatePlan will create a terraform plan.
func Create(account string, statepath string) {
	back := &backend.Local{}
	shell, err := ps.New(back)
	defer shell.Exit()

	_, _, err = shell.Execute("tf plan -out=\"matt/" + account + ".tfplan\" " + statepath)

	if err == nil {
		log.Printf("Created Plan: %v\n", account+".tfplan")
	} else {
		log.Printf("Error: %v\n", err)
	}
}

// tfApplyPlan will apply a terraform plan.
func Apply(account string, statepath string) {
	back := &backend.Local{}
	shell, err := ps.New(back)
	defer shell.Exit()
	stdout, stderr, err := shell.Execute("tf apply -no-color matt/" + account + ".tfplan")

	err = ioutil.WriteFile("matt/"+account+".stdout.log", []byte(stdout), 0)
	err = ioutil.WriteFile("matt/"+account+".stderr.log", []byte(stderr), 0)

	if err == nil {
		log.Printf("Applied Plan: %v\n", account+".tfplan")
	} else {
		_ = ioutil.WriteFile("matt/"+account+".error.log", []byte(err.Error()), 0)
		log.Printf("### ERROR terraforming %v , See logs for details.\n", account)
	}

}

// Destroy will Destroy a terraform plan.
func Destroy(account string, statepath string) {
	back := &backend.Local{}
	shell, err := ps.New(back)
	defer shell.Exit()
	stdout, stderr, err := shell.Execute("tf destroy -auto-approve " + statepath)

	err = ioutil.WriteFile("matt/"+account+".stdout.log", []byte(stdout), 0)
	err = ioutil.WriteFile("matt/"+account+".stderr.log", []byte(stderr), 0)

	if err == nil {
		log.Printf("Deleted: %v\n", account)
	} else {
		_ = ioutil.WriteFile("matt/"+account+".error.log", []byte(err.Error()), 0)
		log.Printf("### ERROR terraforming %v , See logs for details.\n", account)
	}

}

func check(e error) {
	if e != nil {
		log.Printf("Error: %v\n", e)
		return
	}
}
