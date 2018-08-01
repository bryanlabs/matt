package terraform

import (
	"io/ioutil"
	"log"
	"os"

	ps "github.com/bhendo/go-powershell"
	"github.com/bhendo/go-powershell/backend"
)

// tfInit will initialize the new state.
func Init(statepath string, options string, account string, modulename string) {
	back := &backend.Local{}
	shell, err := ps.New(back)

	check(err)

	defer shell.Exit()

	backend := "-backend-config=\"key=" + account + "-" + modulename + ".tfstate\""
	cmd := "tf init -reconfigure " + options + " " + backend + " " + statepath
	// fmt.Println("Init: ", cmd)
	_, _, err = shell.Execute(cmd)
	if err != nil {
		log.Printf(err.Error())
	}

}

// tfCreatePlan will create a terraform plan.
func Create(account string, statepath string, options string, modulename string) {
	os.MkdirAll("matt/plans", os.ModePerm)
	back := &backend.Local{}
	shell, err := ps.New(back)
	defer shell.Exit()

	cmd := "tf plan -no-color -out=\"matt/plans/" + account + "-" + modulename + ".tfplan\" " + options + " " + statepath
	// fmt.Println("Create: ", cmd)
	_, _, err = shell.Execute(cmd)

	if err == nil {
		log.Printf("Created Plan: %v\n", modulename+".tfplan")
	} else {
		log.Printf("Error: %v\n", err)
	}
}

// tfApplyPlan will apply a terraform plan.
func Apply(account string, statepath string, options string, modulename string) {
	back := &backend.Local{}
	shell, err := ps.New(back)
	defer shell.Exit()

	cmd := "tf apply -no-color matt/plans/" + account + "-" + modulename + ".tfplan"
	// fmt.Println("Apply: ", cmd)
	stdout, stderr, err := shell.Execute(cmd)

	err = ioutil.WriteFile("matt/"+account+"-"+modulename+".stdout.log", []byte(stdout), 0)
	if len(stderr) > 0 {
		err = ioutil.WriteFile("matt/"+account+"-"+modulename+".stderr.log", []byte(stderr), 0)
	}

	if err == nil {
		log.Printf("Applied Plan: %v\n", account+"-"+modulename+".tfplan")
	} else {
		_ = ioutil.WriteFile(modulename+".error.log", []byte(err.Error()), 0)
		log.Printf("### ERROR terraforming %v , See logs for details.\n", account)
	}

}

// Destroy will Destroy a terraform plan.
func Destroy(account string, statepath string, options string, modulename string) {
	back := &backend.Local{}
	shell, err := ps.New(back)
	defer shell.Exit()
	cmd := "tf destroy -no-color -auto-approve " + options + " " + statepath
	// fmt.Println("Destroy: ", cmd)
	stdout, stderr, err := shell.Execute(cmd)

	err = ioutil.WriteFile("matt/"+account+"-"+modulename+".stdout.log", []byte(stdout), 0)
	if len(stderr) > 0 {
		err = ioutil.WriteFile("matt/"+account+"-"+modulename+".stderr.log", []byte(stderr), 0)
	}

	if err == nil {
		log.Printf("Deleted: %v\n", account)
	} else {
		_ = ioutil.WriteFile("matt/"+account+"-"+modulename+".error.log", []byte(err.Error()), 0)
		log.Printf("### ERROR terraforming %v , See logs for details.\n", account)
	}

}

func check(e error) {
	if e != nil {
		log.Printf("Error: %v\n", e)
		return
	}
}
