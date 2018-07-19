package terraform

// tfInit will initialize the new state.
func tfInit(statepath string) error {
	back := &backend.Local{}
	shell, err := ps.New(back)

	check(err)

	defer shell.Exit()

	_, _, err = shell.Execute("tf init -reconfigure " + statepath)
	return err
}

// tfCreatePlan will create a terraform plan.
func tfCreatePlan(account string, statepath string) error {
	back := &backend.Local{}
	shell, err := ps.New(back)
	defer shell.Exit()

	_, _, err = shell.Execute("tf plan -out=\"" + account + ".tfplan\" " + statepath)

	if err == nil {
		log.Printf("Created Plan: %v\n", account+".tfplan")
	}

	return err
}

// tfApplyPlan will create a terraform plan.
func tfApplyPlan(account string, statepath string) error {
	back := &backend.Local{}
	shell, err := ps.New(back)
	defer shell.Exit()
	stdout, stderr, err := shell.Execute("tf apply " + account + ".tfplan")

	err = ioutil.WriteFile("logs/std/"+account+".stdout.log", []byte(stdout), 0)
	err = ioutil.WriteFile("logs/std/"+account+".stderr.log", []byte(stderr), 0)

	if err == nil {
		log.Printf("Applied Plan: %v\n", account+".tfplan")
	}

	return err

}