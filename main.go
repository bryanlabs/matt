package main

import (
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	ps "github.com/bhendo/go-powershell"
	"github.com/bhendo/go-powershell/backend"
	"github.com/go-ini/ini"
	"github.com/guregu/dynamo"
)

// The important fields from the AWS Named Profile.
type profile struct {
	name string
	arn  string
}

type dbaccountinfo struct {
	AccountID     string
	Project       string
	IsDevelopment bool
	Deleted_at    string
}

// Loop over all profiles and terraform them.
func main() {
	for _, profile := range getProfiles() {
		err := goTerraform(profile, os.Args[1])
		if err != nil {
			log.Printf("### ERROR terraforming %v , See logs for details.\n", profile.name)
			continue
		}
	}
}

// getProfiles will return a list of valid/enabled named profiles found in a users aws profile.
func getProfiles() []profile {
	xp := make([]profile, 0)

	// Get the userprofile from powershell.
	back := &backend.Local{}
	shell, err := ps.New(back)
	defer shell.Exit()
	check(err)
	stdout, _, err := shell.Execute("$env:userprofile")
	check(err)

	// Load config and remove carriage returns.
	re := regexp.MustCompile(`\r?\n`)
	awsconfig := re.ReplaceAllString(stdout, "") + "\\.aws\\config"
	cfg, err := ini.Load(awsconfig)
	if err != nil {
		log.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}

	// Find Valid/Enabled Named profiles. EG: default profile has no arn.
	for _, section := range cfg.Sections() {
		if section.HasKey("role_arn") {
			var p profile
			p.name = section.Name()
			p.arn = section.Key("role_arn").String()
			xp = append(xp, p)
		}
	}
	//return the named profiles
	log.Printf("Found %v Named Profile(s) in: %v", len(xp), awsconfig)
	return xp
}

// goTerraform will used a named profile to apply a terraform state.
func goTerraform(p profile, statepath string) error {
	log.Printf("Terraforming %v with %v\n", p.name, statepath)

	// get the account number from arn.
	arnslice := strings.Split(p.arn, ":")
	account := arnslice[4]

	//Update account_id for provider and tfvars.
	providerpath := statepath + "/provider.tf"
	updateAccountID(providerpath, account)
	tfvarspath := "vars.auto.tfvars"
	updateAccountID(tfvarspath, account)
	updateAllAccounts()

	// Initialize the new state file
	err := tfInit(statepath)

	// Plan the change.
	if err == nil {
		err = tfCreatePlan(account, statepath)
	}

	if err == nil {
		// Plan the change.
		err = tfApplyPlan(account, statepath)
	}
	if err != nil {
		_ = ioutil.WriteFile("logs/errors/"+account+".error.log", []byte(err.Error()), 0)
		return err
	}
	log.Printf("Terraforming %v Complete\n", p.name)
	return err
}

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

// updateAccountID will modify provider.tf and dynamic.tfvars to allow multi account terraforming.
func updateAccountID(file string, account string) {
	data, err := ioutil.ReadFile(file)
	check(err)
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if strings.Contains(line, "tfstate") {
			lines[i] = "    key    = \"" + account + ".tfstate\""
		} else if strings.HasPrefix(line, "account_id =") {
			lines[i] = "account_id = \"" + account + "\""
		}
	}
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(file, []byte(output), 0)
}

func updateAllAccounts() {
	// Create DynamoDB client

	db := dynamo.New(session.New(), &aws.Config{Region: aws.String("us-east-1")})
	table := db.Table("AccountSecurityGroups")
	var results []dbaccountinfo

	err := table.Scan().All(&results)

	if err != nil {
		log.Fatal("Could not get session: ", err)
	}

	var allAccounts []string
	for _, r := range results {
		if r.Deleted_at == "" && r.IsDevelopment != true {
			allAccounts = append(allAccounts, r.AccountID)
		}
	}

	// Update the tfvars file.
	file := "allaccounts.auto.tfvars"
	var accountARNs []string
	var accountNumbers []string
	for _, account := range allAccounts {
		//  "arn:aws:iam::117751863401:root",
		ARN := "\"arn:aws:sts::" + account + ":assumed-role/lambda-ssmhealth/ssmhealth\""
		accountNumbers = append(accountNumbers, "\""+account+"\"")
		accountARNs = append(accountARNs, ARN)
		// straccountARNs := strings.Join(accountARNs, ",")
	}
	data, err := ioutil.ReadFile(file)
	check(err)
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if strings.Contains(line, "all_accountarns") {
			lines[i] = "all_accountarns = [" + strings.Join(accountARNs, ", ") + "]"
		} else if strings.Contains(line, "all_accountnumbers") {
			lines[i] = "all_accountnumbers = [" + strings.Join(accountNumbers, ", ") + "]"
		}
	}
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(file, []byte(output), 0)
}

func check(e error) {
	if e != nil {
		log.Printf("Error: %v\n", e)
		return
	}
}
