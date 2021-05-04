package dsa

import (
	"fmt"
	"time"
)

// Ticket struct logs fields for the database items:
// TicketID    int        Unique ticket identifier
// Product     int        Input restricted to existing products (managed by admin)
// Status      int        Input restricted to existing statuses (hardcoded)
// Category    int        Input restricted to existing categories (hardcoded)
// Priority    int        Input restricted to existing priority (hardcoded)
// EstHours    int        Estimated hours of work for ticket
// StartDate   time.Time  Input date
// DueDate     time.Time  Due date
// Creator     string     Input restricted to existing usernames
// Title       string     Header summary for ticket
// Description string     Elaboration for ticket
// Assignee    string     Input restricted to existing usernames

type Ticket struct {
	TicketID                                      int64
	Product, Status, Category, Priority, EstHours int
	StartDate, DueDate                            time.Time
	Creator, Title, Description, Assignee         string
}

// TicketNode struct specifies fields for an Ticket struct, two pointers to other Nodes within storage AVL Tree, and height for maintaining balance within AVL tree.
type TicketNode struct {
	Ticket Ticket
	Height int
	Left   *TicketNode
	Right  *TicketNode
}

// blankticket is the return value for Newticket in certain conditions where new ticket submission should not proceed.
var blankticket = Ticket{
	TicketID:    -1,
	Product:     -1,
	Status:      -1,
	Category:    -1,
	EstHours:    -1,
	Priority:    -1,
	StartDate:   time.Now(),
	DueDate:     time.Now(),
	Creator:     "",
	Title:       "",
	Description: "",
	Assignee:    "",
}

// New TicketNode Operations

// Newticket creates a new Ticket (taking user input) and encapsulates it in a TicketNode for processing in priority queue / AVL tree, returning its pointer.
func Newticket(
	ticketID int64,
	product, status, category, priority, estHours int,
	startDate, dueDate time.Time,
	creator, title, description, assignee string,
	priorities, products, statuses, categories *[]string,
) Ticket {
	newticket := Ticket{
		TicketID:    ticketID,
		Product:     product,
		Status:      status,
		Category:    category,
		EstHours:    estHours,
		Priority:    priority,
		StartDate:   startDate,
		DueDate:     dueDate,
		Creator:     creator,
		Title:       title,
		Description: description,
		Assignee:    assignee,
	}

	fmt.Println("Created new ticket.")
	printTicket(newticket, priorities, products, statuses, categories)
	return newticket
}

// // EditTicket edits values at an existing ticket node already in the AVL tree
// func EditTicket(original *TicketNode,
// 	products, statuses, categories, priorities *[]string,
// 	users *[]*UserNode) {

// 	result := original.Ticket

// 	options := &[]string{
// 		"Title",
// 		"Description",
// 		"Assignee",
// 		"Estimated Hours to Complete",
// 		"Priority",
// 		"Due Date",
// 		"Product",
// 		"Status",
// 		"Category",
// 	}

// 	choiceErr := errhand.ErrInvalid
// 	for choiceErr != nil {
// 		printoptions(*options)
// 		optioncount := len(*options)
// 		input, err := menuchoiceExit(optioncount)

// 		if err == nil {
// 			fmt.Println("Selected [", input, "] : ", (*options)[input-1])
// 			switch input {
// 			case 1: // Title
// 				scanner := bufio.NewScanner(os.Stdin)
// 				fmt.Print("Input new ticket title (string): ")
// 				if scanner.Scan() {
// 					fmt.Printf("You inputted \"%s\"\n", scanner.Text())
// 				}
// 				result.Title = scanner.Text()
// 			case 2: // Description
// 				scanner := bufio.NewScanner(os.Stdin)
// 				fmt.Print("Input new ticket description (string): ")
// 				if scanner.Scan() {
// 					fmt.Printf("You inputted \"%s\"\n", scanner.Text())
// 				}
// 				result.Description = scanner.Text()
// 			case 3: // Assignee
// 				assignerr := errhand.ErrInvalid
// 				for assignerr != nil {
// 					PrintHT(users, PrintSLLnoadmin)
// 					fmt.Println("Assign the ticket to an existing user (enter one of the usernames above):")
// 					scanner := bufio.NewScanner(os.Stdin)
// 					fmt.Print("Input ticket assignee (string): ")
// 					if scanner.Scan() {
// 						fmt.Printf("You inputted \"%s\"\n", scanner.Text())
// 					}
// 					newassignee := scanner.Text()

// 					found, _ := HTLookup(users, newassignee)

// 					if found {
// 						result.Assignee = newassignee
// 						assignerr = nil
// 					} else {
// 						fmt.Println(assignerr)
// 					}
// 				}
// 			case 4: // Estimated Hours to Complete
// 				fmt.Println("Editing Estimated Hours:")
// 				esthourserr := errhand.ErrInvalid
// 				var newestHours int
// 				for esthourserr != nil {
// 					fmt.Println("Input new Estimated Hours (integer):")
// 					fmt.Scanln(&newestHours)

// 					if newestHours <= 0 {
// 						fmt.Println(esthourserr)
// 						fmt.Println("Estimated hours cannot be negative!")
// 					} else {
// 						result.EstHours = newestHours
// 						esthourserr = nil
// 					}
// 				}

// 			case 5: // Priority
// 				fmt.Println("Editing Priority:")
// 				prierr := errhand.ErrInvalid
// 				for prierr != nil {
// 					printoptions(*priorities)
// 					pricount := len(*priorities)
// 					input, err := menuchoice(pricount)
// 					if err == nil {
// 						fmt.Println("Selected [", input, "] : ", (*priorities)[input-1])
// 						result.Priority = input
// 						prierr = nil
// 					} else {
// 						fmt.Println(err)
// 					}
// 				}
// 			case 6: // Due Date
// 				// Input is tracked as a duration over and above startDate
// 				var years, months, days int

// 				fmt.Println("Due date will be logged based on adding a given number of years, months and days to the logged start date")
// 				fmt.Println("Input number of years (integer): ")
// 				fmt.Scanln(&years)
// 				fmt.Println("Input number of months (integer): ")
// 				fmt.Scanln(&months)
// 				fmt.Println("Input number of days (integer): ")
// 				fmt.Scanln(&days)

// 				result.DueDate = result.StartDate.AddDate(years, months, days)
// 				fmt.Println("New Due Date logged at", result.DueDate)
// 			case 7: // Product
// 				fmt.Println("Editing Product:")
// 				proderr := errhand.ErrInvalid
// 				for proderr != nil {
// 					printoptions(*products)
// 					productcount := len(*products)
// 					input, err := menuchoice(productcount)
// 					if err == nil {
// 						fmt.Println("Selected [", input, "] : ", (*products)[input-1])
// 						result.Product = input
// 						proderr = nil
// 					} else {
// 						fmt.Println(err)
// 					}
// 				}
// 			case 8: // Status
// 				fmt.Println("Status:")
// 				staterr := errhand.ErrInvalid
// 				for staterr != nil {
// 					printoptions(*statuses)
// 					statuscount := len(*statuses)
// 					input, err := menuchoice(statuscount)
// 					if err == nil {
// 						fmt.Println("Selected [", input, "] : ", (*statuses)[input-1])
// 						result.Status = input
// 						staterr = nil
// 					} else {
// 						fmt.Println(err)
// 					}
// 				}
// 			case 9: // Category
// 				fmt.Println("Editing Category:")
// 				caterr := errhand.ErrInvalid
// 				for caterr != nil {
// 					printoptions(*categories)
// 					catcount := len(*categories)
// 					input, err := menuchoice(catcount)
// 					if err == nil {
// 						fmt.Println("Selected [", input, "] : ", (*categories)[input-1])
// 						result.Category = input
// 						caterr = nil
// 					} else {
// 						fmt.Println(err)
// 					}
// 				}
// 			}
// 			original.Ticket = result
// 			choiceErr = nil
// 		} else if err == errhand.ErrTerminate {
// 			fmt.Println(err)
// 			break
// 		} else {
// 			fmt.Println(err)
// 		}
// 	}
// }

// printTicket returns a formatted print of a ticket, with all appropriate values parsed for passing into the relevant HTML template.
func printTicket(ticket Ticket, priorities, products, statuses, categories *[]string) []string {
	var s []string
	s = append(s, "")
	s = append(s, fmt.Sprintln("Title:", ticket.Title))
	s = append(s, fmt.Sprintln("Description:", ticket.Description))
	s = append(s, fmt.Sprintln(""))
	s = append(s, fmt.Sprintln("Creator:", ticket.Creator))
	s = append(s, fmt.Sprintln("Assignee:", ticket.Assignee))
	s = append(s, fmt.Sprintln(""))
	s = append(s, fmt.Sprintln("Estimated Hours to Complete:", ticket.EstHours))
	s = append(s, fmt.Sprintln("Priority:", (*priorities)[ticket.Priority]))
	s = append(s, fmt.Sprintln("Start Date:", ticket.StartDate))
	s = append(s, fmt.Sprintln("Due Date:", ticket.DueDate))
	s = append(s, fmt.Sprintln(""))
	s = append(s, fmt.Sprintln("Product:", (*products)[ticket.Product]))
	s = append(s, fmt.Sprintln("Status:", (*statuses)[ticket.Status]))
	s = append(s, fmt.Sprintln("Category:", (*categories)[ticket.Category]))
	s = append(s, fmt.Sprintln("Ticket ID:", ticket.TicketID))
	s = append(s, fmt.Sprintln("------------------------------"))

	return s
}
