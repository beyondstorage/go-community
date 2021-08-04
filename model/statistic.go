package model

import "fmt"

// Statistic count events we needed
type Statistic struct {
	PROpened    uint
	PRClosed    uint
	IssueOpened uint
	IssueClosed uint
}

type Statistics []Statistic

func (s Statistic) String() string {
	return fmt.Sprintf("PR open: %d, close: %d; Issue open: %d, close: %d", s.PROpened, s.PRClosed, s.IssueOpened, s.IssueClosed)
}

// FormatPrint format statistic as print needed
func (s Statistic) FormatPrint() string {
	return fmt.Sprintf(`
## Weekly Stats

| | Opened this week | Closed this week |
| ---- | ---- | ---- |
| Issues | %d | %d |
| PR's | %d | %d |
`, s.IssueOpened, s.IssueClosed, s.PROpened, s.PRClosed)
}

// CountPROpen add PROpened counter
func (s *Statistic) CountPROpen() {
	s.PROpened++
}

// CountPRClose add PRClosed counter
func (s *Statistic) CountPRClose() {
	s.PRClosed++
}

// CountIssueOpen add IssueOpened counter
func (s *Statistic) CountIssueOpen() {
	s.IssueOpened++
}

// CountIssueClose add IssueClosed counter
func (s *Statistic) CountIssueClose() {
	s.IssueClosed++
}

// IsBlank check whether the statistic is blank (each field equals 0)
func (s *Statistic) IsBlank() bool {
	return s.PROpened+s.PRClosed+s.IssueOpened+s.IssueClosed == 0
}

// Sum all statistics and return the final result as a Statistic
func (s Statistics) Sum() Statistic {
	res := Statistic{}
	for _, stat := range s {
		res.PROpened += stat.PROpened
		res.PRClosed += stat.PRClosed
		res.IssueOpened += stat.IssueOpened
		res.IssueClosed += stat.IssueClosed
	}
	return res
}
