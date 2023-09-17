package routertwo

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

var router *Router

var errDummy = errors.New("not implemented in DummyRoutable")

func TestRouter(t *testing.T) {
	d, err := os.MkdirTemp("", "remotemoe-router-test")
	if err != nil {
		t.Fatalf("could not create temporary database: %s", err)
	}

	t.Logf("Opening router with database %s", d)
	router, err = NewRouter(d)
	if err != nil {
		t.Fatalf("unable to create new router: %s", err)
	}
}

type DummyRoutable struct {
	Name string
}

func (r *DummyRoutable) FQDN() string {
	if r.Name == "" {
		panic("Dummyroutable cannot have empty name")
	}
	return r.Name
}

func (r *DummyRoutable) DialContext(_ context.Context, _, _ string) (net.Conn, error) {
	return nil, errDummy
}

func (r *DummyRoutable) Replaced() {}

func TestReplace(t *testing.T) {
	dummy := &DummyRoutable{Name: "dummyd.remote.moe"}
	replaced, err := router.Online(dummy)
	if err != nil {
		t.Fatalf("unable to replace first time: %s", err)
	}

	// we should not replace anything the first time
	if replaced != false {
		t.Fatalf("dummy replaced a route even though it shouldn't have")
	}

	replaced, err = router.Online(dummy)
	if err != nil {
		t.Fatalf("unable to replace existing route: %s", err)
	}

	if replaced != true {
		t.Fatalf("dummy did not replace previous dummy")
	}

	// so, we should be able to dial our dummy, and receive an error
	_, err = router.DialContext(context.TODO(), "tcp", "dummyd.remote.moe:80")
	if !errors.Is(err, errDummy) {
		t.Fatalf("we did not expect error: %s", err)
	}
}

func TestClean(t *testing.T) {
	router.Clean(time.Second)
}

func TestRestoreDatabase(t *testing.T) {
	r, err := NewRouter("database_test")
	if err != nil {
		t.Fatalf("unable to restore database: %s", err)
	}

	_, err = r.DialContext(context.TODO(), "tcp", "dummy.remote.moe:80")
	if !errors.Is(err, ErrOffline) {
		t.Fatalf("we expected this host to be offline")
	}

	_, err = r.DialContext(context.TODO(), "tcp", "blarh.remote.moe:80")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("this host should not exist")
	}

	// In this database, there should be a "dummy.remote.moe" host
	_, exists := (*r.active)["dummy.remote.moe"]
	if !exists {
		t.Fatalf("an expected route did not exist")
	}
}

// TestRace checks if multiple routines are able to DialContext
// while getting replaced
func TestRace(t *testing.T) {
	var g errgroup.Group

	router.Online(&DummyRoutable{Name: "race.remote.moe"})

	for i := 0; i < 5; i++ {

		g.Go(func() error {
			_, err := router.DialContext(context.TODO(), "tcp", "race.remote.moe:80")
			if !errors.Is(err, errDummy) {
				return fmt.Errorf("did not expect %w", err)
			}
			return nil
		})
	}

	g.Go(func() error {
		router.Offline(&DummyRoutable{Name: "race.remote.moe"})
		return nil
	})

	err := g.Wait()
	if err != nil {
		t.Fatalf("error: %s", err)
	}
}

func TestAddRemoveName(t *testing.T) {
	rtbl := &DummyRoutable{Name: "addremovename.remote.moe"}
	_, err := router.Online(rtbl)
	if err != nil {
		t.Fatalf("unexpected error from router: %s", err)
	}

	router.Offline(rtbl)
	name := NewName("hejhej.remote.moe", rtbl)

	err = router.AddName(name)
	if err != nil {
		t.Fatalf("could not add name: %s", err)
	}

	// we should now be able to dial the name, and end up with our dummyroute
	_, err = router.DialContext(context.TODO(), "tcp", "hejhej.remote.moe:80")
	if !errors.Is(err, ErrOffline) {
		t.Fatalf("unexpected error from router: %s", err)
	}

	// we should also check if the name appeared on our filesystem
	predictedPath := path.Join(router.dbPath, "hejhej.remote.moe.json")

	_, err = os.Stat(predictedPath)
	if err != nil {
		t.Fatalf("unexpected error from stat: %s", err)
	}

	err = router.RemoveName("hejhej.remote.moe", rtbl)
	if err != nil {
		t.Fatalf("did not expect remove error: %s", err)
	}

	// now if we dial the same name, we should receive a notfound error
	_, err = router.DialContext(context.TODO(), "tcp", "hejhej.remote.moe:80")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("we expected a not found error: %s", err)
	}

	// also, the file should have been removed from the filesystem
	_, err = os.Stat(predictedPath)
	if err == nil {
		t.Fatalf("the file was not removed from fs")
	}
}

func TestIndex(t *testing.T) {
	rtbl := &DummyRoutable{Name: "testindex.remote.moe"}
	name := NewName("hello.remote.moe", rtbl)
	name2 := NewName("hello2.remote.moe", rtbl)

	router.index(name)
	router.index(name2)

	list := router.nameIndex["testindex.remote.moe"]

	t.Logf("index: %+v", router.nameIndex)
	router.reduceIndex("testindex.remote.moe", name)
	t.Logf("index: %+v", router.nameIndex)
	router.reduceIndex("testindex.remote.moe", name2)
	t.Logf("index: %+v", router.nameIndex)
	t.Logf("list: %+v", list)

}

func TestRemoveNames(t *testing.T) {
	rtbl := &DummyRoutable{Name: "testremovename.remote.moe"}
	router.Online(rtbl)

	name := NewName("hello.remote.moe", rtbl)
	name2 := NewName("hello2.remote.moe", rtbl)

	err := router.AddName(name)
	if err != nil {
		t.Fatal(err)
	}

	router.AddName(name2)
	if err != nil {
		t.Fatal(err)
	}

	_, err = router.DialContext(context.TODO(), "tcp", "hello.remote.moe:80")
	if !errors.Is(err, errDummy) {
		t.Fatalf("expected errDummy, got: %s", err)
	}

	_, err = router.DialContext(context.TODO(), "tcp", "hello2.remote.moe:80")
	if !errors.Is(err, errDummy) {
		t.Fatalf("expected errDummy, got: %s", err)
	}

	names, err := router.RemoveNames(rtbl)
	if err != nil {
		t.Fatal(err)
	}

	if names[0].FQDN() != "hello.remote.moe" {
		t.Fatalf("first name was not hello.remote.moe, %s", names[0].FQDN())
	}

	if names[1].FQDN() != "hello2.remote.moe" {
		t.Fatal("second name was not hello2.remote.moe")
	}

	// and now, we should not be able to dial them any more
	_, err = router.DialContext(context.TODO(), "tcp", "hello2.remote.moe:80")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %s", err)
	}

	// and now, we should not be able to dial them any more
	_, err = router.DialContext(context.TODO(), "tcp", "hello.remote.moe:80")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %s", err)
	}
}

func TestNames(t *testing.T) {
	rtbl := &DummyRoutable{Name: "testnames.remote.moe"}
	name := NewName("hello.remote.moe", rtbl)
	name2 := NewName("hello2.remote.moe", rtbl)

	router.AddName(name)
	router.AddName(name2)

	names, err := router.Names(rtbl)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if names[0].FQDN() != "hello.remote.moe" {
		t.Fatalf("unexpected FQDN of first item: %s", names[0].FQDN())
	}

	if names[1].FQDN() != "hello2.remote.moe" {
		t.Fatalf("unexpected FQDN of second item: %s", names[0].FQDN())
	}
}
