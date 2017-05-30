package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	rq "github.com/parnurzeal/gorequest"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var pushDB string

// Uploads a database to a DBHub.io cloud.
var pushCmd = &cobra.Command{
	Use:   "push [database file]",
	Short: "Upload a database",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Ensure a database file was given
		if len(args) == 0 {
			return errors.New("No database file specified")
		}
		// TODO: Allow giving multiple database files on the command line.  Hopefully just needs turning this
		// TODO  into a for loop
		if len(args) > 1 {
			return errors.New("Only one database can be uploaded at a time (for now)")
		}

		// Ensure the database file exists
		file := args[0]
		fi, err := os.Stat(file)
		if err != nil {
			return err
		}

		// Ensure commit message has been provided
		if msg == "" {
			return errors.New("Missing commit message!")
		}

		// Determine name to store database as
		if pushDB == "" {
			pushDB = filepath.Base(file)
		}

		// Send the file
		req := rq.New().Post(cloud+"/db_upload").
			Type("multipart").
			Set("Branch", branch).
			Set("Message", msg).
			Set("ModTime", fi.ModTime().Format(time.RFC3339)).
			Set("Database", pushDB).
			SendFile(file)
		if name != "" && email != "" {
			req.Set("Author", name)
			req.Set("Email", email)
		}
		resp, _, errs := req.End()
		if errs != nil {
			log.Print("Errors when uploading database to the cloud:")
			for _, err := range errs {
				log.Print(err.Error())
			}
			return errors.New("Error when uploading database to the cloud")
		}
		if resp != nil && resp.StatusCode != http.StatusCreated {
			return errors.New(fmt.Sprintf("Upload failed with an error: HTTP status %d - '%v'\n",
				resp.StatusCode, resp.Status))
		}
		fmt.Printf("%s - Database upload successful.  Name: %s, size: %d, branch: %s\n", cloud,
			pushDB, fi.Size(), branch)
		return nil
	},
}

func init() {
	RootCmd.AddCommand(pushCmd)
	pushCmd.Flags().StringVar(&branch, "branch", "master",
		"Remote branch the database will be uploaded to")
	pushCmd.Flags().StringVar(&email, "email", "", "Email address of the author")
	pushCmd.Flags().StringVar(&msg, "message", "",
		"(Required) Commit message for this upload")
	pushCmd.Flags().StringVar(&name, "author", "", "Author name")
	pushCmd.Flags().StringVar(&pushDB, "dbname", "", "Override for the database name")
}
