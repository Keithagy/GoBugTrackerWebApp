package dsa

import (
	"strings"
)

// AVLtree is a struct describing AVL trees as used in this application.
// It comprises a pointer to the AVL tree root, as well as the sorting criteria (out of TicketNode fields) used to build the tree.
type AVLtree struct {
	Root     *TicketNode
	Sortfunc func(newticket *TicketNode, junction *TicketNode) bool
}

// NewAVLT creates a new AVLTree specified with some sortfunc, and returns its pointer.
func NewAVLT(sortfunc func(newticket *TicketNode, junction *TicketNode) bool) *AVLtree {
	return &AVLtree{nil, sortfunc}
}

// IOtraversal implements an in-order traversal of an already-existing AVL tree, for which the root pointer is passed as the first argument.
func IOtraversal(avlroot *TicketNode, priorities, products, statuses, categories *[]string, result [][]string) [][]string {
	if avlroot == nil {
		return result
	}

	result = IOtraversal(avlroot.Left, priorities, products, statuses, categories, result)
	result = append(result, printTicket(avlroot.Ticket, priorities, products, statuses, categories))
	result = IOtraversal(avlroot.Right, priorities, products, statuses, categories, result)
	return result
}

// AVLinsert recursively inserts a node root of a specified subtree, does required rotations.
// For valid sortfuncs to use as arguments, see section below on AVLtree sortfuncs.
// Returns new root of the subtree.
func AVLinsert(newticket *TicketNode,
	sortfunc func(newticket *TicketNode, junction *TicketNode) bool,
	subtree *TicketNode) *TicketNode {

	// BST Insertion
	if subtree == nil {
		newticket.Height = 1
		return (newticket)
	}

	if sortfunc(newticket, subtree) {
		subtree.Left = AVLinsert(newticket, sortfunc, subtree.Left)
	} else {
		subtree.Right = AVLinsert(newticket, sortfunc, subtree.Right)
	}

	// Update heights of parent nodes
	subtree.Height = 1 + max(height(subtree.Left), height(subtree.Right))

	// Check if parent nodes balanced
	bf := getbf(subtree)

	// Left Left
	if bf > 1 && sortfunc(newticket, subtree.Left) {
		return rightrotate(subtree)
	}

	// Right Right
	if bf < -1 && !sortfunc(newticket, subtree.Right) {
		return leftrotate(subtree)
	}

	// Left Right
	if bf > 1 && !sortfunc(newticket, subtree.Left) {
		subtree.Left = leftrotate(subtree.Left)
		return rightrotate(subtree)
	}

	// Right Left
	if bf < -1 && sortfunc(newticket, subtree.Right) {
		subtree.Right = rightrotate(subtree.Right)
		return leftrotate(subtree)
	}

	// No change needed because balanced
	return subtree
}

// Mytickets traverses an AVL tree (at a given root node), and returns pointer to a subsetted AVL tree containing only nodes with a particular username as creator.
func Mytickets(avlroot, result *TicketNode,
	sortfunc func(newticket *TicketNode, junction *TicketNode) bool,
	user string) *TicketNode {
	if avlroot == nil {
		return result
	} else if avlroot.Ticket.Creator == user {
		copied := &TicketNode{
			Ticket: avlroot.Ticket,
			Height: 0,
			Left:   nil,
			Right:  nil,
		}
		result = AVLinsert(copied, sortfunc, result)
		result = Mytickets(avlroot.Left, result, sortfunc, user)
		result = Mytickets(avlroot.Right, result, sortfunc, user)
	} else {
		result = Mytickets(avlroot.Left, result, sortfunc, user)
		result = Mytickets(avlroot.Right, result, sortfunc, user)
	}
	return result
}

// Myassigns traverses an AVL tree (at a given root node), and returns pointer to a subsetted AVL tree containing only nodes assigned to a particular username.
func Myassigns(avlroot, result *TicketNode,
	sortfunc func(newticket *TicketNode, junction *TicketNode) bool,
	user string) *TicketNode {
	if avlroot == nil {
		return result
	} else if avlroot.Ticket.Assignee == user {
		copied := &TicketNode{
			Ticket: avlroot.Ticket,
			Height: 0,
			Left:   nil,
			Right:  nil,
		}
		result = AVLinsert(copied, sortfunc, result)
		result = Myassigns(avlroot.Left, result, sortfunc, user)
		result = Myassigns(avlroot.Right, result, sortfunc, user)
	} else {
		result = Myassigns(avlroot.Left, result, sortfunc, user)
		result = Myassigns(avlroot.Right, result, sortfunc, user)
	}
	return result
}

// AVLpivot takes an existing AVL tree and re-sorts it using a different sorting function. Returns a pointer to the re-sorted AVL tree.
// For valid sortfuncs to use as arguments, see section below on AVLtree sortfuncs.
func AVLpivot(source, destination *TicketNode,
	sortfunc func(newticket *TicketNode, junction *TicketNode) bool) *TicketNode {
	if source == nil {
	} else {
		copied := &TicketNode{
			Ticket: source.Ticket,
			Height: 0,
			Left:   nil,
			Right:  nil,
		}
		destination = AVLpivot(source.Left, destination, sortfunc)
		destination = AVLinsert(copied, sortfunc, destination)
		destination = AVLpivot(source.Right, destination, sortfunc)
	}
	return destination
}

// AVLdelete recursively deletes a target node (with a certain TicketID).
// Returns root of the modified subtree.
// Important: Assumes tree is sorted by ticketID.
func AVLdelete(subtree *TicketNode, targetID int64) *TicketNode {
	if subtree == nil {
		return subtree
	}

	// Does root node contain key to be deleted? If not,
	// Does key to be deleted lie to the left or right subtree?
	if avlSearchDirect(subtree, targetID) == -1 {
		subtree.Left = AVLdelete(subtree.Left, targetID)
	} else if avlSearchDirect(subtree, targetID) == 1 {
		subtree.Right = AVLdelete(subtree.Right, targetID)
	} else {
		var tmp *TicketNode

		// Node with one subtree or less
		if subtree.Left == nil || subtree.Right == nil {
			if subtree.Left == nil {
				tmp = subtree.Left
			} else {
				tmp = subtree.Right
			}

			if tmp == nil { // Node with no subtrees
				tmp = subtree
				subtree = nil
			} else { // Node with one subtree
				*subtree = *tmp
			}
		} else {
			// Node with two subtrees (get inorder successor, smallest in right subtree)
			tmp = MinIDnode(subtree.Right)

			// Copy inorder successor's data to this node
			subtree.Ticket = tmp.Ticket

			// Delete inorder successor
			subtree.Right = AVLdelete(subtree.Right, tmp.Ticket.TicketID)
		}
	}

	// If tree only has one lone node, return
	if subtree == nil {
		return subtree
	}

	// Update height of current node
	subtree.Height = 1 + max(height(subtree.Left), height(subtree.Right))

	// Check if parent nodes balanced
	bf := getbf(subtree)

	// Left Left
	if bf > 1 && getbf(subtree.Left) >= 0 {
		return rightrotate(subtree)
	}

	// Right Right
	if bf < -1 && getbf(subtree.Right) <= 0 {
		return leftrotate(subtree)
	}

	// Left Right
	if bf > 1 && getbf(subtree.Left) < 0 {
		subtree.Left = leftrotate(subtree.Left)
		return rightrotate(subtree)
	}

	// Right Left
	if bf < -1 && getbf(subtree.Right) > 0 {
		subtree.Right = rightrotate(subtree.Right)
		return leftrotate(subtree)
	}

	// No change needed because balanced
	return subtree
}

// AVLsearch searches the AVLtree (sorted by ticketID only) and returns to pointer to that node, if it exists (otherwise, returns nil).
func AVLsearch(subtree *TicketNode, targetID int64) *TicketNode {
	if subtree == nil {
		return nil
	} else if avlSearchDirect(subtree, targetID) == -1 {
		subtree = AVLsearch(subtree.Left, targetID)
	} else if avlSearchDirect(subtree, targetID) == 1 {
		subtree = AVLsearch(subtree.Right, targetID)
	} else {
		return subtree
	}
	return subtree // Implicitly only gets here if targetID not contined in tree
}

// AVL tree sortfuncs;
// Returns bool for program flow control; true and false direct tree traversal/recursion left and right, respectively

// ByTicketID
func ByTicketID(newticket *TicketNode, junction *TicketNode) bool {
	return (newticket.Ticket.TicketID < junction.Ticket.TicketID)
}

// ByProduct -- uses ByTicket ID as tiebreaker
func ByProduct(newticket *TicketNode, junction *TicketNode) bool {
	if newticket.Ticket.Product < junction.Ticket.Product {
		return true
	} else if newticket.Ticket.Product == junction.Ticket.Product {
		return ByTicketID(newticket, junction)
	}
	return false
}

// ByStatus -- uses ByTicket ID as tiebreaker
func ByStatus(newticket *TicketNode, junction *TicketNode) bool {
	if newticket.Ticket.Status < junction.Ticket.Status {
		return true
	} else if newticket.Ticket.Status == junction.Ticket.Status {
		return ByTicketID(newticket, junction)
	}
	return false
}

// ByCategory -- uses ByTicket ID as tiebreaker
func ByCategory(newticket *TicketNode, junction *TicketNode) bool {
	if newticket.Ticket.Category < junction.Ticket.Category {
		return true
	} else if newticket.Ticket.Category == junction.Ticket.Category {
		return ByTicketID(newticket, junction)
	}
	return false
}

// ByPriority -- uses ByTicket ID as tiebreaker
func ByPriority(newticket *TicketNode, junction *TicketNode) bool {
	if newticket.Ticket.Priority < junction.Ticket.Priority {
		return true
	} else if newticket.Ticket.Priority == junction.Ticket.Priority {
		return ByTicketID(newticket, junction)
	}
	return false
}

// ByEstHours -- uses ByTicket ID as tiebreaker
func ByEstHours(newticket *TicketNode, junction *TicketNode) bool {
	if newticket.Ticket.EstHours < junction.Ticket.EstHours {
		return true
	} else if newticket.Ticket.EstHours == junction.Ticket.EstHours {
		return ByTicketID(newticket, junction)
	}
	return false
}

// ByStartDate -- uses ByTicket ID as tiebreaker
func ByStartDate(newticket *TicketNode, junction *TicketNode) bool {
	if newticket.Ticket.StartDate.Before(junction.Ticket.StartDate) {
		return true
	} else if newticket.Ticket.StartDate == junction.Ticket.StartDate {
		return ByTicketID(newticket, junction)
	}
	return false
}

// ByDueDate -- uses ByTicket ID as tiebreaker
func ByDueDate(newticket *TicketNode, junction *TicketNode) bool {
	if newticket.Ticket.DueDate.Before(junction.Ticket.DueDate) {
		return true
	} else if newticket.Ticket.DueDate == junction.Ticket.DueDate {
		return ByTicketID(newticket, junction)
	}
	return false
}

// ByCreator -- uses ByTicket ID as tiebreaker
func ByCreator(newticket *TicketNode, junction *TicketNode) bool {
	if strings.ToLower(newticket.Ticket.Creator) < strings.ToLower(junction.Ticket.Creator) {
		return true
	} else if strings.EqualFold(newticket.Ticket.Creator, junction.Ticket.Creator) {
		return ByTicketID(newticket, junction)
	}
	return false
}

// ByTitle -- uses ByTicket ID as tiebreaker
func ByTitle(newticket *TicketNode, junction *TicketNode) bool {
	if strings.ToLower(newticket.Ticket.Title) < strings.ToLower(junction.Ticket.Title) {
		return true
	} else if strings.EqualFold(newticket.Ticket.Title, junction.Ticket.Title) {
		return ByTicketID(newticket, junction)
	}
	return false
}

// ByDescription -- uses ByTicket ID as tiebreaker
func ByDescription(newticket *TicketNode, junction *TicketNode) bool {
	if strings.ToLower(newticket.Ticket.Description) < strings.ToLower(junction.Ticket.Description) {
		return true
	} else if strings.EqualFold(newticket.Ticket.Description, junction.Ticket.Description) {
		return ByTicketID(newticket, junction)
	}
	return false
}

// ByAssignee -- uses ByTicket ID as tiebreaker
func ByAssignee(newticket *TicketNode, junction *TicketNode) bool {
	if strings.ToLower(newticket.Ticket.Assignee) < strings.ToLower(junction.Ticket.Assignee) {
		return true
	} else if strings.EqualFold(newticket.Ticket.Assignee, junction.Ticket.Assignee) {
		return ByTicketID(newticket, junction)
	}
	return false
}

// Utility functions

// avlSearchDirect is a helper function for tree search / deletion;
// Returns 1 for right subtree, 0 if target found, -1 for left subtree
func avlSearchDirect(subtree *TicketNode, targetID int64) int {
	if targetID < subtree.Ticket.TicketID {
		return -1
	} else if targetID == subtree.Ticket.TicketID {
		return 0
	}
	return 1
}

// Calculates the height of the AVL subtree with a given root.
func height(tree *TicketNode) int {
	if tree == nil {
		return 0
	}

	lheight := height(tree.Left)
	rheight := height(tree.Right)

	if lheight > rheight {
		return (lheight + 1)
	}
	return (rheight + 1)
}

// Returns max of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Rotates subtree at a given root right
func rightrotate(subtree *TicketNode) *TicketNode {
	leftsub := subtree.Left
	orphan := leftsub.Right

	leftsub.Right = subtree
	subtree.Left = orphan

	subtree.Height = max(height(subtree.Left), height(subtree.Right)+1)
	leftsub.Height = max(height(leftsub.Left), height(leftsub.Right)+1)

	return leftsub
}

// Rotates subtree at a given root left
func leftrotate(subtree *TicketNode) *TicketNode {
	rightsub := subtree.Right
	orphan := rightsub.Left

	rightsub.Left = subtree
	subtree.Right = orphan

	subtree.Height = max(height(subtree.Left), height(subtree.Right)+1)
	rightsub.Height = max(height(rightsub.Left), height(rightsub.Right)+1)

	return rightsub
}

// Calculate balance factor for the root of a given subtree
func getbf(subtree *TicketNode) int {
	if subtree == nil {
		return 0
	}
	return height(subtree.Left) - height(subtree.Right)
}

// MinIDnode returns the node within a subtree with the minimum key value in that tree
func MinIDnode(subtree *TicketNode) *TicketNode {
	current := subtree
	for current.Left != nil {
		current = current.Left
	}

	return current

}
