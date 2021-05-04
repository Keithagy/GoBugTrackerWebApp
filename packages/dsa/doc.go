/*
   ticket.go:
   Ticket is a custom struct containing all the fields to be tracked, pivoted, edited and displayed.
   It underlies both the priority queue and AVL tree data structures implemented in the application.
   For storage and handling of the data across function calls, Tickets are contained within TicketNodes, which contain other fields required to store Tickets within the ticket log AVL tree.
   A Ticket can only be created by a non-admin user. When a non-admin user creates a ticket, it is not directly added into the ticket log AVL tree; rather, it is first added to a priority queue called submissions, which is implemented via an array-based heap.
   From submissions, the ticket must first be approved by an admin user before the ticket is wrapped in a TicketNode and added to the ticket log AVL tree.

   avltree.go:
   Implements an adapted AVL tree and associated functions to initialize, traverse, insert nodes, delete nodes and others.
   Each AVLtree contains a pointer to the root node (of type TicketNode), as well as a specification of the sorting criteria applied by the AVLtree:
   - Ticket ID
   - Product
   - Status
   - Category
   - Priority
   - Estimated Hours to Complete
   - Start Date
   - Due Date
   - Ticket Creator
   - Ticket Title
   - Ticket Description
   - Ticket Assignee
   AVLtrees can be pivoted to apply a different sorting criteria, changing the order in which tickets are displayed.

   heap.go:
   Array-based implementation of a heap, which is used as a priority queue used to track user submissions.
   Note that in this heap, the topmost tickets are the highest-priority ones (lowest priority score).
   Thus, an admin appoving user ticket submissions will first start from the highest-priority ticket, making use of the min-heap property.

   userhash.go:
   Implements a hash table, used in the application to record and manipulate information of user accounts in-memory.
   Hash table is implemented as an array of SLLs made up of UserNodes.
   UserNodes contain a User, and a pointer to another UserNode (thus forming the SLL).
   User info includes a username(string), bcrypt-hashed password([]byte), and admin status(bool)
*/
package dsa
