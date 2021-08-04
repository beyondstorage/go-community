package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatistic(t *testing.T) {
	stat := Statistic{}
	assert.True(t, stat.IsBlank())

	stat.CountIssueOpen()
	assert.Equal(t, uint(1), stat.IssueOpened, "after issue open count")

	stat.CountIssueClose()
	assert.Equal(t, uint(1), stat.IssueClosed, "after issue close count")

	stat.CountPROpen()
	assert.Equal(t, uint(1), stat.PROpened, "after pr open count")

	stat.CountPRClose()
	assert.Equal(t, uint(1), stat.PRClosed, "after pr close count")

	list := make([]Statistic, 0)
	list = append(list, stat, Statistic{
		PROpened:    3,
		PRClosed:    4,
		IssueOpened: 5,
		IssueClosed: 6,
	})

	assert.Equal(t, Statistic{
		PROpened:    4,
		PRClosed:    5,
		IssueOpened: 6,
		IssueClosed: 7,
	}, Statistics(list).Sum(), "after statistics sum together")
}
