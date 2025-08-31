package search

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"unicode"
)

const (
	KindRepositories = "repositories"
	KindCode         = "code"
	KindIssues       = "issues"
	KindCommits      = "commits"
)

type Query struct {
	Keywords   []string
	Kind       string
	Limit      int
	Order      string
	Page       int
	Qualifiers Qualifiers
	Sort       string
}

type Qualifiers struct {
	Archived            *bool
	Assignee            string
	Author              string
	AuthorDate          string
	AuthorEmail         string
	AuthorName          string
	Base                string
	Closed              string
	Commenter           string
	Comments            string
	Committer           string
	CommitterDate       string
	CommitterEmail      string
	CommitterName       string
	Created             string
	Draft               *bool
	Extension           string
	Filename            string
	Followers           string
	Fork                string
	Forks               string
	GoodFirstIssues     string
	Hash                string
	Head                string
	HelpWantedIssues    string
	In                  []string
	Interactions        string
	Involves            string
	Is                  []string
	Label               []string
	Language            string
	License             []string
	Mentions            string
	Merge               *bool
	Merged              string
	Milestone           string
	No                  []string
	Parent              string
	Path                string
	Project             string
	Pushed              string
	Reactions           string
	Repo                []string
	Review              string
	ReviewRequested     string
	ReviewedBy          string
	Size                string
	Stars               string
	State               string
	Status              string
	Team                string
	TeamReviewRequested string
	Topic               []string
	Topics              string
	Tree                string
	Type                string
	Updated             string
	User                []string
}

// String returns the string representation of the query which can be used with
// search API (except for advanced issue search), global search GUI (i.e.
// github.com/search), or Pull Requests tab (in repositories).
func (q Query) String() string {
	qualifiers := formatQualifiers(q.Qualifiers, nil)
	keywords := formatKeywords(q.Keywords)
	all := append(keywords, qualifiers...)
	return strings.TrimSpace(strings.Join(all, " "))
}

// AdvancedIssueSearchString returns the string representation of the query
// compatible with the advanced issue search syntax. The query can be used in
// Issues tab (of repositories) and the Issues dashboard (i.e.
// github.com/issues).
func (q Query) AdvancedIssueSearchString() string {
	qualifiers := formatQualifiers(q.Qualifiers, formatAdvancedIssueSearch)
	keywords := formatKeywords(q.Keywords)
	all := append(keywords, qualifiers...)
	return strings.TrimSpace(strings.Join(all, " "))
}

func formatAdvancedIssueSearch(qualifier string, vs []string) (s []string, applicable bool) {
	switch qualifier {
	case "in":
		return formatSpecialQualifiers("in", vs, []string{"title", "body", "comments"}), true
	case "is":
		return formatSpecialQualifiers("is", vs, []string{"private", "public"}), true
	case "user", "repo":
		return []string{groupWithOR(qualifier, vs)}, true
	}
	// Let the default formatting take over
	return nil, false
}

func formatSpecialQualifiers(qualifier string, vs []string, valuesToOR []string) []string {
	specials := make([]string, 0, len(vs))
	rest := make([]string, 0, len(vs))
	for _, v := range vs {
		if slices.Contains(valuesToOR, v) {
			specials = append(specials, v)
		} else {
			rest = append(rest, v)
		}
	}

	all := make([]string, 0, 2)
	if len(specials) > 0 {
		all = append(all, groupWithOR(qualifier, specials))
	}
	if len(rest) > 0 {
		for _, v := range rest {
			all = append(all, fmt.Sprintf("%s:%s", qualifier, quote(v)))
		}
	}

	slices.Sort(all)
	return all
}

func groupWithOR(qualifier string, vs []string) string {
	if len(vs) == 0 {
		return ""
	}

	all := make([]string, 0, len(vs))
	for _, v := range vs {
		all = append(all, fmt.Sprintf("%s:%s", qualifier, quote(v)))
	}

	if len(all) == 1 {
		return all[0]
	}

	slices.Sort(all)
	return fmt.Sprintf("(%s)", strings.Join(all, " OR "))
}

func (q Qualifiers) Map() map[string][]string {
	m := map[string][]string{}
	v := reflect.ValueOf(q)
	t := reflect.TypeOf(q)
	for i := 0; i < v.NumField(); i++ {
		fieldName := t.Field(i).Name
		key := camelToKebab(fieldName)
		typ := v.FieldByName(fieldName).Kind()
		value := v.FieldByName(fieldName)
		switch typ {
		case reflect.Ptr:
			if value.IsNil() {
				continue
			}
			v := reflect.Indirect(value)
			m[key] = []string{fmt.Sprintf("%v", v)}
		case reflect.Slice:
			if value.IsNil() {
				continue
			}
			s := []string{}
			for i := 0; i < value.Len(); i++ {
				if value.Index(i).IsZero() {
					continue
				}
				s = append(s, fmt.Sprintf("%v", value.Index(i)))
			}
			m[key] = s
		default:
			if value.IsZero() {
				continue
			}
			m[key] = []string{fmt.Sprintf("%v", value)}
		}
	}
	return m
}

func quote(s string) string {
	if strings.ContainsAny(s, " \"\t\r\n") {
		return fmt.Sprintf("%q", s)
	}
	return s
}

// formatQualifiers renders qualifiers into a plain query.
//
// The formatter is a custom formatting function that can be used to modify the
// output of each qualifier. If the formatter returns (nil, false) the default
// formatting will be applied.
func formatQualifiers(qs Qualifiers, formatter func(qualifier string, vs []string) (s []string, applicable bool)) []string {
	type entry struct {
		key    string
		values []string
	}

	var all []entry
	for k, vs := range qs.Map() {
		if len(vs) == 0 {
			continue
		}

		e := entry{key: k}

		if formatter != nil {
			if s, applicable := formatter(k, vs); applicable {
				e.values = s
				all = append(all, e)
				continue
			}
		}

		for _, v := range vs {
			e.values = append(e.values, fmt.Sprintf("%s:%s", k, quote(v)))
		}
		if len(e.values) > 1 {
			slices.Sort(e.values)
		}
		all = append(all, e)
	}

	slices.SortFunc(all, func(a, b entry) int {
		return strings.Compare(a.key, b.key)
	})

	result := make([]string, 0, len(all))
	for _, e := range all {
		result = append(result, e.values...)
	}
	return result
}

func formatKeywords(ks []string) []string {
	result := make([]string, len(ks))
	for i, k := range ks {
		before, after, found := strings.Cut(k, ":")
		if !found {
			result[i] = quote(k)
		} else {
			result[i] = fmt.Sprintf("%s:%s", before, quote(after))
		}
	}
	return result
}

// CamelToKebab returns a copy of the string s that is converted from camel case form to '-' separated form.
func camelToKebab(s string) string {
	var output []rune
	var segment []rune
	for _, r := range s {
		if !unicode.IsLower(r) && string(r) != "-" && !unicode.IsNumber(r) {
			output = addSegment(output, segment)
			segment = nil
		}
		segment = append(segment, unicode.ToLower(r))
	}
	output = addSegment(output, segment)
	return string(output)
}

func addSegment(inrune, segment []rune) []rune {
	if len(segment) == 0 {
		return inrune
	}
	if len(inrune) != 0 {
		inrune = append(inrune, '-')
	}
	inrune = append(inrune, segment...)
	return inrune
}
