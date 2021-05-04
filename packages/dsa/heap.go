package dsa

import "fmt"

// Heap Operations

// Addsubmission inserts a new node to the priority queue.
func Addsubmission(submissions *[]Ticket, newticket Ticket) {
	// Insert new ticket at end of submissions
	*submissions = append(*submissions, newticket)

	// Ensure heap property is maintained
	for index := len(*submissions) - 1; index != 0 && (*submissions)[parent(index)].Priority > (*submissions)[index].Priority; index = parent(index) {
		swap(&((*submissions)[index]), &((*submissions)[parent(index)]))
	}
}

// Searchsubmissions for a given node in priority queue. Returns true if the ticket exists, alongside the target ticket ID.
func Searchsubmissions(submissions *[]Ticket, ticketID int64) (bool, int64) {
	for index := int64(0); index < int64(len(*submissions)); index++ {
		if (*submissions)[index].TicketID == ticketID {
			return true, index
		}
	}
	return false, -1
}

// Popsubmission removes the root node of the priority queue, returning the removed node.
func Popsubmission(submissions *[]Ticket) Ticket {
	popped := (*submissions)[0]
	(*submissions)[0] = (*submissions)[len(*submissions)-1]
	(*submissions) = (*submissions)[:len(*submissions)-1]

	Makeheap(submissions, 0)

	return popped
}

// LOtraversal implements in-order traversal of the heap.
func LOtraversal(submissions *[]Ticket, priorities, products, statuses, categories *[]string) [][]string {
	var s [][]string
	if len(*submissions) == 0 {
		fmt.Println("No submissions outstanding.")
	} else {
		for index := 0; index < len(*submissions); index++ {
			s = append(s, printTicket((*submissions)[index], priorities, products, statuses, categories))
		}
	}
	return s
}

// Utility function for swapping elements of priority queue in-place.
func swap(x *Ticket, y *Ticket) {
	tmp := *x
	*x = *y
	*y = tmp
}

// Utility function which returns the node with the lower priority value (more important). If both have equal priority values, the function returns the node with the sooner due date.
func firstinline(x *Ticket, y *Ticket) *Ticket {
	if (*x).Priority < (*y).Priority {
		return x
	} else if (*x).Priority > (*y).Priority {
		return y
	} else {
		if ((*x).DueDate).Before((*y).DueDate) {
			return x
		}
		return y
	}
}

// Utility function returning indices of parent and children nodes in heap.
func parent(child int) int {
	return (child - 1) / 2
}

func left(parent int) int {
	return (2*parent + 1)
}

func right(parent int) int {
	return (2*parent + 2)
}

// Makeheap preserves the heap property of the priority queue.
func Makeheap(submissions *[]Ticket, root int) {
	if right(root) < len(*submissions) {
		toheapify := firstinline(&(*submissions)[right(root)], &(*submissions)[left(root)])
		swapped := false
		if firstinline(&(*submissions)[root], &(*submissions)[left(root)]) == &(*submissions)[left(root)] || firstinline(&(*submissions)[root], &(*submissions)[right(root)]) == &(*submissions)[right(root)] {
			swap(&(*submissions)[root], toheapify)
			swapped = true
		}

		if swapped {
			if toheapify == &(*submissions)[left(root)] {
				Makeheap(submissions, left(root))
			} else {
				Makeheap(submissions, right(root))
			}
		}
	}
}
