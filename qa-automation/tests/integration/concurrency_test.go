package integration

import (
	"sync"
	"testing"

	. "moneyflow/qa-automation/helpers"
)

type result struct {
	status int
	code   string
	err    error
}

func runParallel(n int, fn func(i int) (Response, error)) []result {
	res := make([]result, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			r, err := fn(i)
			res[i] = result{status: r.Status, code: r.ErrorCode(), err: err}
		}(i)
	}
	wg.Wait()
	return res
}

func TestTC_TRF_018_ConcurrentTransfersDoNotOverdraw(t *testing.T) {
	const workers, amount = 50, 1000
	a := CreateAccount(t, "RUB", 10000)
	b := CreateAccount(t, "RUB", 0)

	res := runParallel(workers, func(int) (Response, error) {
		return SendTransfer(NewKey(), a.ID, b.ID, amount)
	})

	ok, insufficient := 0, 0
	for _, r := range res {
		switch {
		case r.err != nil:
			t.Errorf("request error: %v", r.err)
		case r.status == 201:
			ok++
		case r.status == 422 && r.code == "INSUFFICIENT_FUNDS":
			insufficient++
		default:
			t.Errorf("unexpected response: %d %s", r.status, r.code)
		}
	}
	if ok != 10 || insufficient != 40 {
		t.Fatalf("got %d ok and %d insufficient, want 10 and 40", ok, insufficient)
	}
	AssertBalance(t, a.ID, 0)
	AssertBalance(t, b.ID, 10000)
}

func TestTC_TRF_019_ConcurrentSameIdempotencyKey(t *testing.T) {
	const workers = 20
	a := CreateAccount(t, "RUB", 100000)
	b := CreateAccount(t, "RUB", 0)
	key := NewKey()

	res := runParallel(workers, func(int) (Response, error) {
		return SendTransfer(key, a.ID, b.ID, 1000)
	})

	created, replayed := 0, 0
	for _, r := range res {
		switch {
		case r.err != nil:
			t.Errorf("request error: %v", r.err)
		case r.status == 201:
			created++
		case r.status == 200:
			replayed++
		default:
			t.Errorf("unexpected response: %d %s", r.status, r.code)
		}
	}
	if created != 1 || replayed != workers-1 {
		t.Fatalf("got %d created and %d replayed, want 1 and %d", created, replayed, workers-1)
	}
	AssertBalance(t, a.ID, 99000)
	AssertBalance(t, b.ID, 1000)
}

func TestTC_TRF_020_OppositeTransfersNoDeadlock(t *testing.T) {
	const workers = 50
	a := CreateAccount(t, "RUB", 100000)
	b := CreateAccount(t, "RUB", 100000)

	res := runParallel(workers, func(i int) (Response, error) {
		if i%2 == 0 {
			return SendTransfer(NewKey(), a.ID, b.ID, 100)
		}
		return SendTransfer(NewKey(), b.ID, a.ID, 100)
	})

	for _, r := range res {
		if r.err != nil {
			t.Errorf("request error: %v", r.err)
		} else if r.status != 201 {
			t.Errorf("unexpected response: %d %s", r.status, r.code)
		}
	}
	if sum := Balance(t, a.ID) + Balance(t, b.ID); sum != 200000 {
		t.Fatalf("sum of balances: got %d, want 200000", sum)
	}
}