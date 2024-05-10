package server_test

import (
	"testing"

	"github.com/Shucck/go-stratum-server/testutil"
)

func TestStratumServer(t *testing.T) {
	t.Run("TestSubscription", testSubscription)
	t.Run("TestAuthorization", testAuthorization)
	t.Run("TestSetDifficulty", testSetDifficulty)
	t.Run("TestSendJob", testSendJob)
	t.Run("TestProcessSolution", testProcessSolution)
}

func testSubscription(t *testing.T) {
	l := testutil.StartServer(t)
	defer l.Close()

	conn := testutil.ConnectServer(t, l)
	defer conn.Close()

	testutil.SendRequest(t, conn, map[string]interface{}{
		"id":     1,
		"method": "mining.subscribe",
	})

	testutil.WaitShort()
}

func testAuthorization(t *testing.T) {
	l := testutil.StartServer(t)
	defer l.Close()

	conn := testutil.ConnectServer(t, l)
	defer conn.Close()

	testutil.SendRequest(t, conn, map[string]interface{}{
		"id":     1,
		"method": "mining.authorize",
		"params": []interface{}{"test_user", "test_password"},
	})

	testutil.WaitShort()
}

func testSetDifficulty(t *testing.T) {
	l := testutil.StartServer(t)
	defer l.Close()

	conn := testutil.ConnectServer(t, l)
	defer conn.Close()

	testutil.SendRequest(t, conn, map[string]interface{}{
		"id":     1,
		"method": "mining.set_difficulty",
		"params": []interface{}{5.0},
	})

	testutil.WaitShort()
}

func testSendJob(t *testing.T) {
	l := testutil.StartServer(t)
	defer l.Close()

	conn := testutil.ConnectServer(t, l)
	defer conn.Close()

	testutil.SendRequest(t, conn, map[string]interface{}{
		"id":     1,
		"method": "mining.notify",
		"params": []interface{}{"job_id", "prevhash", "coinb1", "coinb2", []interface{}{"merkle_branches"}, "version", "nbits", "ntime", true},
	})

	testutil.WaitShort()
}

func testProcessSolution(t *testing.T) {
	l := testutil.StartServer(t)
	defer l.Close()

	conn := testutil.ConnectServer(t, l)
	defer conn.Close()

	testutil.SendRequest(t, conn, map[string]interface{}{
		"id":     1,
		"method": "mining.submit",
		"params": []interface{}{"test_miner", "job_id", "nonce", "result"},
	})

	testutil.WaitShort()
}
