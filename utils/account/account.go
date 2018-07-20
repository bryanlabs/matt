package account

import (
	"io/ioutil"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
)

type AWS_dbaccountinfo struct {
	AccountID     string
	Project       string
	IsDevelopment bool
	Deleted_at    string
}

// updateAccountID will modify provider.tf and dynamic.tfvars to allow multi account terraforming.
func UpdateAccountID(file string, account string) {
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

func GetEvoAccounts() (evoaccounts []string) {
	// Create DynamoDB client

	db := dynamo.New(session.New(), &aws.Config{Region: aws.String("us-east-1")})
	table := db.Table("AccountSecurityGroups")
	var results []AWS_dbaccountinfo

	err := table.Scan().All(&results)

	if err != nil {
		log.Fatal("Could not get session: ", err)
	}

	for _, r := range results {
		if r.Deleted_at == "" && r.IsDevelopment != true {
			evoaccounts = append(evoaccounts, r.AccountID)
		}
	}
	return
}

func UpdateAllAccounts(allAccounts []string) {

	// Update the tfvars file.
	file := "vars.auto.tfvars"
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
