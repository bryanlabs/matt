package auth

import (
	"log"
	"os"
	"regexp"

	ps "github.com/bhendo/go-powershell"
	"github.com/bhendo/go-powershell/backend"
	"github.com/go-ini/ini"
)

// The important fields from the AWS Named Profile.
type AWS_Named_Profile struct {
	name string
	arn  string
}

// getProfiles will return a list of valid/enabled named profiles found in a users aws profile.
func getProfiles() []AWS_Named_Profile {
	xp := make([]AWS_Named_Profile, 0)

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
			var p AWS_Named_Profile
			p.name = section.Name()
			p.arn = section.Key("role_arn").String()
			xp = append(xp, p)
		}
	}
	//return the named profiles
	log.Printf("Found %v Named Profile(s) in: %v", len(xp), awsconfig)
	return xp
}

func check(e error) {
	if e != nil {
		log.Printf("Error: %v\n", e)
		return
	}
}
