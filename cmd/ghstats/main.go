package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v50/github"
)

// Model
type Package struct {
	FullName      string
	Description   string
	StarsCount    int
	ForksCount    int
	LastUpdatedBy string
}

func main() {
	token := os.Getenv("GH_API_TOKEN")

	var r, o string
	var repositories []string
	var resp *github.Response
	var err error
	flag.StringVar(&o, "o", "", "organisation name")
	flag.StringVar(&r, "r", "", "repository list separed by comma")
	flag.Parse()

	if r == "" {
		panic(fmt.Errorf("repository list can not be empty"))
	}
	if o == "" {
		panic(fmt.Errorf("organisation can not be empty"))
	}
	ctx := context.Background()
	client := github.NewTokenClient(ctx, token)
	listOption := github.ListOptions{
		Page:    0,
		PerPage: 200,
	}

	csvFilename := "./repositories-stats.csv"
	fc, err := os.Create(csvFilename)
	if err != nil {
		panic(fmt.Errorf("error while creating in file %s: %w", csvFilename, err))
	}
	defer fc.Close()

	// csv headers
	_, err = fc.Write([]byte(fmt.Sprintf("repository, pr, state, login, draft, createdAt, mergedAt, prDuration (d), submittedAt, reviewer, rewiewState, reviewDelay (d)\n")))
	if err != nil {
		panic(fmt.Errorf("error while writing in file %s: %w", csvFilename, err))
	}

	repositories = strings.Split(r, ",")
	if err != nil {
		panic(fmt.Errorf("error while splitting %s: %w", r, err))
	}

	for _, repository := range repositories {
		var prs []*github.PullRequest
		prs, resp, err = client.PullRequests.List(ctx,
			o,
			repository,
			&github.PullRequestListOptions{
				State:       "all",
				ListOptions: listOption,
			})
		if err != nil {
			fmt.Printf("\nerror: %v\n", err)
			return
		}

		for _, pr := range prs {
			var rvs []*github.PullRequestReview

			_, err := fc.Write([]byte(fmt.Sprintf("%s, %d, %s, %s, %t, ", repository, *pr.Number, *pr.State, *pr.User.Login, *pr.Draft)))
			if err != nil {
				panic(fmt.Errorf("error while writing in file %s: %w", csvFilename, err))
			}
			switch *pr.State {
			case "closed":
				if pr.MergedAt != nil {
					d := pr.MergedAt.Time.Sub(pr.CreatedAt.Time)
					_, err := fc.Write([]byte(fmt.Sprintf("%s, %s, %.02f", pr.CreatedAt.Time.Format("2006-01-02"), pr.MergedAt.Time.Format("2006-01-02"), d.Hours()/24)))
					if err != nil {
						panic(fmt.Errorf("error while writing in file %s: %w", csvFilename, err))
					}
				} else {
					d := time.Now().Sub(pr.CreatedAt.Time)
					_, err := fc.Write([]byte(fmt.Sprintf("%s, n/a, %.02f", pr.CreatedAt.Time.Format("2006-01-02"), d.Hours()/24)))
					if err != nil {
						panic(fmt.Errorf("error while writing in file %s: %w", csvFilename, err))
					}
				}
			case "open":
				d := time.Now().Sub(pr.CreatedAt.Time)
				_, err := fc.Write([]byte(fmt.Sprintf("%s, n/a, %.02f", pr.CreatedAt.Time.Format("2006-01-02"), d.Hours()/24)))
				if err != nil {
					panic(fmt.Errorf("error while writing in file %s: %w", csvFilename, err))
				}
			}
			rvs, resp, err = client.PullRequests.ListReviews(ctx,
				"Contentsquare",
				repository,
				*pr.Number,
				&listOption,
			)
			if err != nil {
				fmt.Printf("\nerror: %v\n", err)
				return
			}
			for _, rv := range rvs {
				d := rv.SubmittedAt.Time.Sub(pr.CreatedAt.Time)

				_, err = fc.Write([]byte(fmt.Sprintf(", %s, %s, %s, %.02f", rv.SubmittedAt.Time.Format("2006-01-02"), *rv.User.Login, *rv.State, d.Hours()/24)))
				if err != nil {
					panic(fmt.Errorf("error while writing in file %s: %w", csvFilename, err))
				}
			}
			_, err = fc.Write([]byte(fmt.Sprintln("")))
			if err != nil {
				panic(fmt.Errorf("error while writing in file %s: %w", csvFilename, err))
			}
		}
	}

	// Rate.Limit should most likely be 5000 when authorized.
	log.Printf("Rate: %#v\n", resp.Rate)

}
