// Contains code to populate the demo state for the application; otherwise does not impact running of the application.
package test

import (
	dsa "goInAction2/assignment/packages/dsa"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Testdata is a struct used to contain submissions, users, ticketlog, and products which would be populated when demo mode is activated.
type Testdata struct {
	Testsubs      *[]dsa.Ticket
	Testusers     *[]*dsa.UserNode
	Testticketlog *dsa.AVLtree
	Testproducts  *[]string
}

// Demomode populates data structures with demo mode data.
func Demomode(ticketIDcounter *int64, testdataloader chan Testdata, wg *sync.WaitGroup) {
	defer wg.Done()

	testload := Testdata{
		Testsubs:      nil,
		Testusers:     nil,
		Testticketlog: nil,
		Testproducts:  nil,
	}

	testload.Testsubs = &[]dsa.Ticket{}
	testload.Testusers = dsa.NewHT()
	testload.Testticketlog = dsa.NewAVLT(dsa.ByTicketID)
	testload.Testproducts = &([]string{})

	// Users: 1 admin user, 2 test users
	pw, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.MinCost)
	adminuser := dsa.User{
		Name:  "admin",
		Pw:    pw,
		Admin: true,
	}

	pw, _ = bcrypt.GenerateFromPassword([]byte("user1"), bcrypt.MinCost)
	user1 := dsa.User{
		Name:  "user1",
		Pw:    pw,
		Admin: false,
	}

	pw, _ = bcrypt.GenerateFromPassword([]byte("user2"), bcrypt.MinCost)
	user2 := dsa.User{
		Name:  "user2",
		Pw:    pw,
		Admin: false,
	}

	dsa.AddUser(testload.Testusers, adminuser)
	dsa.AddUser(testload.Testusers, user1)
	dsa.AddUser(testload.Testusers, user2)

	// Products: Flying Saucer, Magic Wand, Arc Reactor
	*(testload.Testproducts) = []string{
		"Flying Saucer",
		"Magic Wand",
		"Arc Reactor",
	}

	// 4 Logged tickets, 2 from each test user
	t := time.Now()
	ticket1 := dsa.Ticket{
		TicketID:    *ticketIDcounter,
		Product:     0,
		Status:      2,
		Category:    1,
		EstHours:    20,
		Priority:    2,
		StartDate:   t,
		DueDate:     t.AddDate(0, 6, 3),
		Creator:     "user1",
		Title:       "Address windscreen frosting",
		Description: "How do they do it on planes?",
		Assignee:    "user1",
	}
	atomic.AddInt64(ticketIDcounter, 1)

	t = time.Now()
	ticket2 := dsa.Ticket{
		TicketID:    *ticketIDcounter,
		Product:     1,
		Status:      0,
		Category:    0,
		EstHours:    65,
		Priority:    1,
		StartDate:   t,
		DueDate:     t.AddDate(2, 1, 1),
		Creator:     "user2",
		Title:       "Make up more spells",
		Description: "Use Harry Potter for reference",
		Assignee:    "user1",
	}
	atomic.AddInt64(ticketIDcounter, 1)

	t = time.Now()
	ticket3 := dsa.Ticket{
		TicketID:    *ticketIDcounter,
		Product:     1,
		Status:      0,
		Category:    2,
		EstHours:    45,
		Priority:    2,
		StartDate:   t,
		DueDate:     t.AddDate(0, 5, 0),
		Creator:     "user1",
		Title:       "Replace core with unicorn mane",
		Description: "Shinier sparks, demonstrated success with child user demographic",
		Assignee:    "user2",
	}
	atomic.AddInt64(ticketIDcounter, 1)

	t = time.Now()
	ticket4 := dsa.Ticket{
		TicketID:    *ticketIDcounter,
		Product:     0,
		Status:      2,
		Category:    0,
		EstHours:    45,
		Priority:    1,
		StartDate:   t,
		DueDate:     t.AddDate(0, 6, 0),
		Creator:     "user2",
		Title:       "Add tractor beam",
		Description: "Food supplies running low, more cows needed",
		Assignee:    "user2",
	}
	atomic.AddInt64(ticketIDcounter, 1)

	t1node := &dsa.TicketNode{
		Ticket: ticket1,
		Height: 0,
		Left:   nil,
		Right:  nil,
	}
	testload.Testticketlog.Root = dsa.AVLinsert(t1node, dsa.ByTicketID, testload.Testticketlog.Root)

	t2node := &dsa.TicketNode{
		Ticket: ticket2,
		Height: 0,
		Left:   nil,
		Right:  nil,
	}
	testload.Testticketlog.Root = dsa.AVLinsert(t2node, dsa.ByTicketID, testload.Testticketlog.Root)

	t3node := &dsa.TicketNode{
		Ticket: ticket3,
		Height: 0,
		Left:   nil,
		Right:  nil,
	}
	testload.Testticketlog.Root = dsa.AVLinsert(t3node, dsa.ByTicketID, testload.Testticketlog.Root)

	t4node := &dsa.TicketNode{
		Ticket: ticket4,
		Height: 0,
		Left:   nil,
		Right:  nil,
	}
	testload.Testticketlog.Root = dsa.AVLinsert(t4node, dsa.ByTicketID, testload.Testticketlog.Root)

	// 4 Submissions, 2 from each test user
	t = time.Now()
	sub1 := dsa.Ticket{
		TicketID:    *ticketIDcounter,
		Product:     0,
		Status:      0,
		Category:    0,
		EstHours:    100,
		Priority:    0,
		StartDate:   t,
		DueDate:     t.AddDate(0, 6, 3),
		Creator:     "user1",
		Title:       "Design convertible tires",
		Description: "Wheel spokes become turbines",
		Assignee:    "user2",
	}
	atomic.AddInt64(ticketIDcounter, 1)

	t = time.Now()
	sub2 := dsa.Ticket{
		TicketID:    *ticketIDcounter,
		Product:     1,
		Status:      1,
		Category:    1,
		EstHours:    10,
		Priority:    2,
		StartDate:   t,
		DueDate:     t.AddDate(1, 0, 2),
		Creator:     "user1",
		Title:       "Refine handguard design",
		Description: "More dragonskin!",
		Assignee:    "user1",
	}
	atomic.AddInt64(ticketIDcounter, 1)

	t = time.Now()
	sub3 := dsa.Ticket{
		TicketID:    *ticketIDcounter,
		Product:     2,
		Status:      1,
		Category:    1,
		EstHours:    20,
		Priority:    1,
		StartDate:   t,
		DueDate:     t.AddDate(1, 0, 2),
		Creator:     "user2",
		Title:       "Miniaturize",
		Description: "Tony Stark was able to build this! In a cave!! With a box of scraps!!!",
		Assignee:    "user1",
	}
	atomic.AddInt64(ticketIDcounter, 1)

	t = time.Now()
	sub4 := dsa.Ticket{
		TicketID:    *ticketIDcounter,
		Product:     0,
		Status:      2,
		Category:    1,
		EstHours:    80,
		Priority:    2,
		StartDate:   t,
		DueDate:     t.AddDate(2, 3, 4),
		Creator:     "user2",
		Title:       "Redesign Airlock",
		Description: "Currently fails above 50,000m",
		Assignee:    "user2",
	}
	atomic.AddInt64(ticketIDcounter, 1)

	dsa.Addsubmission(testload.Testsubs, sub1)
	dsa.Addsubmission(testload.Testsubs, sub2)
	dsa.Addsubmission(testload.Testsubs, sub3)
	dsa.Addsubmission(testload.Testsubs, sub4)

	testdataloader <- testload
}
