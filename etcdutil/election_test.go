package etcdutil_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/mailgun/holster/etcdutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestElection(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	election, err := etcdutil.NewElection(ctx, client, etcdutil.ElectionConfig{
		EventObserver: func(e etcdutil.Event) {
			if e.Err != nil {
				t.Fatal(e.Err.Error())
			}
		},
		Election:  "/my-election",
		Candidate: "me",
	})
	require.Nil(t, err)

	assert.Equal(t, true, election.IsLeader())
	election.Close()
	assert.Equal(t, false, election.IsLeader())
}

func TestTwoCampaigns(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	logrus.SetLevel(logrus.DebugLevel)

	c1, err := etcdutil.NewElection(ctx, client, etcdutil.ElectionConfig{
		EventObserver: func(e etcdutil.Event) {
			if e.Err != nil {
				t.Fatal(e.Err.Error())
			}
		},
		Election:  "/my-election",
		Candidate: "c1",
	})
	require.Nil(t, err)

	c2Chan := make(chan bool, 5)
	c2, err := etcdutil.NewElection(ctx, client, etcdutil.ElectionConfig{
		EventObserver: func(e etcdutil.Event) {
			fmt.Printf("Observed: %t err: %v\n", e.IsLeader, err)
			if err != nil {
				t.Fatal(err.Error())
			}
			c2Chan <- e.IsLeader
		},
		Election:  "/my-election",
		Candidate: "c2",
	})
	require.Nil(t, err)

	assert.Equal(t, true, c1.IsLeader())
	assert.Equal(t, false, c2.IsLeader())

	// Cancel first candidate
	c1.Close()
	assert.Equal(t, false, c1.IsLeader())

	// Second campaign should become leader
	/*assert.Equal(t, false, <-c2Chan)
	assert.Equal(t, true, <-c2Chan)

	c2.Close()
	assert.Equal(t, false, <-c2Chan)*/
}
