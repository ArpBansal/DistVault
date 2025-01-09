DATE - 07-01-2025

Context - Authentication of servers

Decision - Mutual Authentication using Mutual TLS approach - mTLS
Authentication work lies on primary/master node at the current time.
It should be a synchronized most importantly a Atomic Transaction like process.

Ponder over - Whether on change on primary/master node we need to authenticate the existing nodes of cluster.

		Also check if Raft lib provides enough control to do the above process, before finalising this.

Need To be Done, Not using:
	Secure storing of certificates seperately by user.
	May hashicorp/vault be a solution.