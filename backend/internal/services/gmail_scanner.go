package services

import (
	"context"
	"fmt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	gmail "google.golang.org/api/gmail/v1"
	"net/mail"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// ---------- Types ----------

type EmailJobEvent struct {
	MessageID   string    `json:"messageId"`
	Subject     string    `json:"subject"`
	Snippet     string    `json:"snippet"`
	Company     string    `json:"company,omitempty"`
	Title       string    `json:"title,omitempty"`
	Status      string    `json:"status"` // wishlist|applied|interviewing|offer|rejected|withdrawn
	AppliedDate time.Time `json:"appliedDate,omitempty"`
	Source      string    `json:"source"`         // gmail
	Link        string    `json:"link,omitempty"` // direct gmail link
}

type GmailScanner struct{}

func NewGmailScanner() *GmailScanner { return &GmailScanner{} }

type ScanResult struct {
	Events        []EmailJobEvent `json:"events"`
	NextPageToken string          `json:"nextPageToken,omitempty"`
}

// ---------- Queries ----------

var applicationQueries = []string{
	`subject:"application received"`,
	`subject:"thanks for applying"`,
	`subject:"we received your application"`,
	`subject:"your application to"`,
	`subject:"applied for"`,
	`from:jobs-lever.co`,
	`from:greenhouse.io`,
	`from:workday.com`,
	`from:smartrecruiters.com`,
	`from:ashbyhq.com`,
	`from:workable.com`,
	`from:indeed.com`,
	`from:linkedin.com`,
}

var rejectionQueries = []string{
	`subject:"we regret"`,
	`subject:"unfortunately"`,
	`subject:"not moving forward"`,
	`subject:"no longer being considered"`,
	`subject:"pursue other candidates"`,
	`subject:"not selected"`,
}

func joined(qs []string) string { return "(" + strings.Join(qs, " OR ") + ")" }

// ---------- Extractors ----------

var (
	reAtCompany = regexp.MustCompile(`(?i)\bat\s+([A-Za-z0-9&\.\-'\s]+)`)
	companyRes  = []*regexp.Regexp{
		regexp.MustCompile(`(?i)your application to ([\w\s\.\-&']+)`),
		regexp.MustCompile(`(?i)application received (?:at|from) ([\w\s\.\-&']+)`),
		regexp.MustCompile(`(?i)thanks for applying to ([\w\s\.\-&']+)`),
	}
	titleRes = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\bfor (?:the )?position of ([\w\s\.\-/&']+?)\b`),
		regexp.MustCompile(`(?i)\bapplication (?:for|to) ([\w\s\.\-/&']+?) (?:at|with)\b`),
		regexp.MustCompile(`(?i)\byour application to .*? for ([\w\s\.\-/&']+)$`),
		regexp.MustCompile(`(?i)[“"]([^”"]+)[”"]\s*:`), // “Software Engineer”:
		regexp.MustCompile(`(?i)^\s*([^:]+?)\s*:\s*`),  // Software Engineer:
	}
	rejectionIndicators = []string{
		"not moving forward", "unfortunately", "no longer being considered",
		"not selected", "pursue other candidates", "we regret", "regret to inform",
	}
)

var titler = cases.Title(language.AmericanEnglish)

func extractCompany(subject, from string) string {
	s := strings.TrimSpace(subject)
	for _, re := range companyRes {
		if m := re.FindStringSubmatch(s); len(m) > 1 {
			return strings.TrimSpace(m[1])
		}
	}
	if m := reAtCompany.FindStringSubmatch(s); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	// fallback: from-domain first label
	if i := strings.Index(from, "@"); i != -1 {
		d := from[i+1:]
		d = strings.ToLower(strings.TrimSpace(d))
		d = strings.TrimPrefix(d, "mail.")
		if j := strings.Index(d, ">"); j != -1 {
			d = d[:j]
		}
		if k := strings.Index(d, "."); k != -1 {
			// Use the new, recommended title-casing method.
			return titler.String(d[:k])
		}
	}
	return ""
}

func extractTitle(subject string) string {
	s := strings.TrimSpace(subject)
	for _, re := range titleRes {
		if m := re.FindStringSubmatch(s); len(m) > 1 {
			return strings.TrimSpace(m[1])
		}
	}
	if m := regexp.MustCompile(`(?i)^["“]?([^"”]+?)["”]?\s+at\s+`).FindStringSubmatch(s); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return ""
}

// ---------- Paged Scan ----------

// only: "all" | "applied" | "rejected"
// internal/services/gmail_scanner.go

// ... (keep everything above this function the same)

// only: "all" | "applied" | "rejected"
//
//	func (s *GmailScanner) ScanPage(ctx context.Context, srv *gmail.Service, since, until time.Time, max int64, pageToken, only string, existingIDs map[string]struct{}) (ScanResult, error) {
//		if max <= 0 {
//			max = 200
//		}
//		if max > 500 {
//			max = 500 // Gmail list page cap
//		}
//
//		dateFilter := ""
//		if !since.IsZero() {
//			dateFilter += fmt.Sprintf(" after:%s", since.UTC().Format("2006/01/02"))
//		}
//		// Add the 'before' filter for the 'until' date
//		if !until.IsZero() {
//			// Add one day to 'until' to make the range inclusive for that day
//			dateFilter += fmt.Sprintf(" before:%s", until.AddDate(0, 0, 1).UTC().Format("2006/01/02"))
//		}
//
//		var q string
//		switch strings.ToLower(strings.TrimSpace(only)) {
//		case "rejected":
//			q = joined(rejectionQueries)
//		case "applied":
//			q = joined(applicationQueries)
//		default:
//			q = joined(applicationQueries) + " OR " + joined(rejectionQueries)
//		}
//		q = q + dateFilter
//
//		list := srv.Users.Messages.List("me").Q(q).MaxResults(max)
//		if pageToken != "" {
//			list.PageToken(pageToken)
//		}
//		res, err := list.Do()
//		if err != nil {
//			return ScanResult{}, err
//		}
//
//		// concurrent metadata gets (throttle to be nice to API)
//		type one struct {
//			ev  EmailJobEvent
//			ok  bool
//			err error
//		}
//		ch := make(chan one, len(res.Messages))
//		var wg sync.WaitGroup
//		sem := make(chan struct{}, 16) // 16 concurrent gets
//
//		for _, m := range res.Messages {
//			wg.Add(1)
//			go func(id string) {
//				defer wg.Done()
//				sem <- struct{}{}
//				defer func() { <-sem }()
//
//				msg, err := srv.Users.Messages.Get("me", id).
//					Format("metadata").
//					MetadataHeaders("Subject", "Date", "From").
//					Do()
//				if err != nil {
//					ch <- one{err: err}
//					return
//				}
//
//				var subj, from, dateStr string
//				for _, h := range msg.Payload.Headers {
//					switch h.Name {
//					case "Subject":
//						subj = h.Value
//					case "From":
//						from = h.Value
//					case "Date":
//						dateStr = h.Value
//					}
//				}
//
//				// Prefer InternalDate; fallback to parsed Date header
//				var applied time.Time
//				if msg.InternalDate > 0 {
//					applied = time.UnixMilli(msg.InternalDate)
//				} else if dateStr != "" {
//					if t, e := mail.ParseDate(dateStr); e == nil {
//						applied = t
//					}
//				}
//
//				ev := EmailJobEvent{
//					MessageID:   msg.Id,
//					Subject:     subj,
//					Snippet:     msg.Snippet,
//					Company:     extractCompany(subj, from),
//					Title:       extractTitle(subj),
//					Status:      "applied", // default; we’ll upgrade to rejected if needed
//					AppliedDate: applied,
//					Source:      "gmail",
//					Link:        "https://mail.google.com/mail/u/0/#all/" + msg.Id,
//				}
//
//				low := strings.ToLower(subj + " " + msg.Snippet)
//				for _, kw := range rejectionIndicators {
//					if strings.Contains(low, kw) {
//						ev.Status = "rejected"
//						break
//					}
//				}
//				ch <- one{ev: ev, ok: true}
//			}(m.Id)
//		}
//
//		wg.Wait()
//		close(ch)
//		//
//		// out := make([]EmailJobEvent, 0, len(res.Messages))
//		// for x := range ch {
//		// 	if x.ok {
//		// 		out = append(out, x.ev)
//		// 	}
//		// }
//		//
//		// sort.SliceStable(out, func(i, j int) bool { return out[i].AppliedDate.After(out[j].AppliedDate) })
//		// return ScanResult{Events: out, NextPageToken: res.NextPageToken}, nil
//		// --- THIS SECTION IS UPDATED ---
//		unfilteredEvents := make([]EmailJobEvent, 0, len(res.Messages))
//		for x := range ch {
//			if x.ok {
//				unfilteredEvents = append(unfilteredEvents, x.ev)
//			}
//		}
//
//		// NEW: Filter out emails that are already in our database.
//		filteredEvents := make([]EmailJobEvent, 0, len(unfilteredEvents))
//		for _, event := range unfilteredEvents {
//			if _, exists := existingIDs[event.MessageID]; !exists {
//				filteredEvents = append(filteredEvents, event)
//			}
//		}
//
//		// Sort the final, filtered list of new events.
//		sort.SliceStable(filteredEvents, func(i, j int) bool { return filteredEvents[i].AppliedDate.After(filteredEvents[j].AppliedDate) })
//
//		return ScanResult{Events: filteredEvents, NextPageToken: res.NextPageToken}, nil
//	}
func (s *GmailScanner) ScanPage(ctx context.Context, srv *gmail.Service, since, until time.Time, max int64, pageToken, only string, existingIDs map[string]struct{}) (ScanResult, error) {
	if max <= 0 {
		max = 200
	}
	if max > 500 {
		max = 500 // Gmail list page cap
	}

	dateFilter := ""
	if !since.IsZero() {
		dateFilter += fmt.Sprintf(" after:%s", since.UTC().Format("2006/01/02"))
	}
	if !until.IsZero() {
		dateFilter += fmt.Sprintf(" before:%s", until.AddDate(0, 0, 1).UTC().Format("2006/01/02"))
	}

	var q string
	switch strings.ToLower(strings.TrimSpace(only)) {
	case "rejected":
		q = joined(rejectionQueries)
	case "applied":
		q = joined(applicationQueries)
	default:
		q = joined(applicationQueries) + " OR " + joined(rejectionQueries)
	}
	q = q + dateFilter

	list := srv.Users.Messages.List("me").Q(q).MaxResults(max)
	if pageToken != "" {
		list.PageToken(pageToken)
	}
	res, err := list.Do()
	if err != nil {
		return ScanResult{}, err
	}

	type one struct {
		ev  EmailJobEvent
		ok  bool
		err error
	}
	ch := make(chan one, len(res.Messages))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 16)

	for _, m := range res.Messages {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			// FIRST: Check if we already have this ID before making an API call.
			if _, exists := existingIDs[id]; exists {
				ch <- one{ok: false} // Send a signal to skip this one.
				return
			}

			// If it's a new ID, proceed to fetch its details.
			msg, err := srv.Users.Messages.Get("me", id).
				Format("metadata").
				MetadataHeaders("Subject", "Date", "From").
				Do()
			if err != nil {
				ch <- one{err: err}
				return
			}

			// ... (rest of the message parsing logic remains the same)
			var subj, from, dateStr string
			for _, h := range msg.Payload.Headers {
				switch h.Name {
				case "Subject":
					subj = h.Value
				case "From":
					from = h.Value
				case "Date":
					dateStr = h.Value
				}
			}

			var applied time.Time
			if msg.InternalDate > 0 {
				applied = time.UnixMilli(msg.InternalDate)
			} else if dateStr != "" {
				if t, e := mail.ParseDate(dateStr); e == nil {
					applied = t
				}
			}

			ev := EmailJobEvent{
				MessageID:   msg.Id,
				Subject:     subj,
				Snippet:     msg.Snippet,
				Company:     extractCompany(subj, from),
				Title:       extractTitle(subj),
				Status:      "applied",
				AppliedDate: applied,
				Source:      "gmail",
				Link:        "https://mail.google.com/mail/u/0/#all/" + msg.Id,
			}

			low := strings.ToLower(subj + " " + msg.Snippet)
			for _, kw := range rejectionIndicators {
				if strings.Contains(low, kw) {
					ev.Status = "rejected"
					break
				}
			}
			ch <- one{ev: ev, ok: true}
		}(m.Id)
	}

	wg.Wait()
	close(ch)

	out := make([]EmailJobEvent, 0, len(res.Messages))
	for x := range ch {
		if x.ok {
			out = append(out, x.ev)
		}
	}

	sort.SliceStable(out, func(i, j int) bool { return out[i].AppliedDate.After(out[j].AppliedDate) })
	return ScanResult{Events: out, NextPageToken: res.NextPageToken}, nil
}
