package projections

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/retro-framework/go-retro/events"
	"github.com/retro-framework/go-retro/framework/depot"
	"github.com/retro-framework/go-retro/framework/retro"
)

type Session struct {
	ap *Profile
}

func (s Session) IsAnonymous() bool {
	return (s.ap == nil)
}

func (s Session) Profile() Profile {
	if s.ap != nil {
		return *s.ap
	} else {
		return Profile{}
	}
}

type Sessions interface {
	Get(context.Context, retro.PartitionName) Session
}

func NewSessions(p Profiles, evManifest retro.EventManifest, d retro.Depot) *memorySessions {
	return &memorySessions{
		profiles:   p,
		evManifest: evManifest,
		d:          d,
		s:          make(map[retro.PartitionName]retro.URN),
	}
}

type memorySessions struct {
	sync.Mutex
	s map[retro.PartitionName]retro.URN

	profiles   Profiles
	evManifest retro.EventManifest
	d          retro.Depot
}

// Get returns empty sessions unless that session has an associated ID
func (ms *memorySessions) Get(ctx context.Context, pn retro.PartitionName) Session {
	fmt.Println("in session get with ", pn)
	fmt.Printf("sessions mapping %#v\n", ms.s)
	if identityName, ok := ms.s[pn]; ok {
		fmt.Println("trying to get a profile by the name of identityName", identityName)
		profile, err := ms.profiles.Get(ctx, identityName)
		if err != nil {
			panic(err)
			return Session{}
		}
		return Session{&profile}
	}
	return Session{}
}

func (ms *memorySessions) Run(ctx context.Context) {

	var everything = ms.d.Watch(ctx, "*")

	for {
		partitionEvents, err := everything.Next(ctx)
		if err == depot.Done {
			continue
		}
		if err != nil {
			fmt.Println("err on pi", err)
			return
		}
		if partitionEvents != nil {
			go func(evIter retro.EventIterator) {
				for {
					var pEv, err = evIter.Next(ctx)
					if err == depot.Done {
						continue
					}
					if err != nil {
						fmt.Println("sessions: err", err)
						return
					}
					if pEv != nil {

						ev, err := ms.evManifest.ForName(pEv.Name())
						if err != nil {
							fmt.Println("err looking up event", err)
							continue
						}

						err = json.Unmarshal(pEv.Bytes(), &ev)
						if err != nil {
							fmt.Println("err unmarshalling event", err, pEv.CheckpointHash().String())
							continue
						}

						switch pEv.PartitionName().Dirname() {
						case "identity":
							// fmt.Println("sessions projection doing something with an identity")
							// fmt.Println(pEv.PartitionName(), pEv.Name(), fmt.Sprintf("%.80s", string(pEv.Bytes())))
						case "session":
							switch tEv := ev.(type) {
							case *events.AssociateIdentity:
								ms.Lock()
								if tEv.Identity != nil {
									ms.s[pEv.PartitionName()] = tEv.Identity.URN()
								}
								ms.Unlock()
							}
						default:
							continue
						}

					}
				}
			}(partitionEvents)
		}
	}
}
