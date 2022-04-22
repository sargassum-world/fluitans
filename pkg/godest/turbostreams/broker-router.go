//nolint:funlen,gocritic,gocognit,gocyclo // All of this is copied from github.com/labstack/echo
package turbostreams

// Node

type kind uint8

const (
	staticKind kind = iota
	paramKind
	anyKind

	paramLabel = byte(':')
	anyLabel   = byte('*')
)

type node struct {
	kind           kind
	label          byte
	prefix         string
	parent         *node
	staticChildren children
	ppath          string
	pnames         []string
	methodHandler  *methodHandler
	paramChild     *node
	anyChild       *node
	isLeaf         bool
	isHandler      bool
}
type children []*node

func newNode(
	t kind, pre string, p *node, sc children, mh *methodHandler, ppath string, pnames []string,
	paramChildren, anyChildren *node,
) *node {
	// Copied from github.com/labstack/echo's newNode function
	return &node{
		kind:           t,
		label:          pre[0],
		prefix:         pre,
		parent:         p,
		staticChildren: sc,
		ppath:          ppath,
		pnames:         pnames,
		methodHandler:  mh,
		paramChild:     paramChildren,
		anyChild:       anyChildren,
		isLeaf:         sc == nil && paramChildren == nil && anyChildren == nil,
		isHandler:      mh.isHandler(),
	}
}

func (n *node) addStaticChild(c *node) {
	// Copied from github.com/labstack/echo's node.addStaticChild method
	n.staticChildren = append(n.staticChildren, c)
}

func (n *node) findStaticChild(l byte) *node {
	// Copied from github.com/labstack/echo's node.findStaticChild method
	for _, c := range n.staticChildren {
		if c.label == l {
			return c
		}
	}
	return nil
}

func (n *node) findChildWithLabel(l byte) *node {
	// Copied from github.com/labstack/echo's node.findChildWithLabel method
	for _, c := range n.staticChildren {
		if c.label == l {
			return c
		}
	}
	if l == paramLabel {
		return n.paramChild
	}
	if l == anyLabel {
		return n.anyChild
	}
	return nil
}

func (n *node) addHandler(method string, h HandlerFunc) {
	switch method {
	case MethodPub:
		n.methodHandler.pub = h
	case MethodSub:
		n.methodHandler.sub = h
	case MethodUnsub:
		n.methodHandler.unsub = h
	case MethodMsg:
		n.methodHandler.msg = h
	}

	if h != nil {
		n.isHandler = true
	} else {
		n.isHandler = n.methodHandler.isHandler()
	}
}

func (n *node) findHandler(method string) HandlerFunc {
	switch method {
	default:
		return nil
	case MethodPub:
		return n.methodHandler.pub
	case MethodSub:
		return n.methodHandler.sub
	case MethodUnsub:
		return n.methodHandler.unsub
	case MethodMsg:
		return n.methodHandler.msg
	}
}

func (n *node) isLeafNode() bool {
	return n.staticChildren == nil && n.paramChild == nil && n.anyChild == nil
}

// Route

// Route contains a handler and information for matching against requests.
type Route struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Name   string `json:"name"`
}

// Router

// router is the registry of routes for subscription matching and topic path parameter parsing.
type router struct {
	tree   *node
	routes map[string]*Route
	broker *Broker
}

func newRouter(b *Broker) *router {
	return &router{
		tree: &node{
			methodHandler: new(methodHandler),
		},
		routes: make(map[string]*Route),
		broker: b,
	}
}

func (r *router) Add(method, path string, h HandlerFunc) {
	// Copied from github.com/labstack/echo's Router.Add method
	// Validate path
	if path == "" {
		path = "/"
	}
	if path[0] != '/' {
		path = "/" + path
	}
	pnames := []string{} // Param names
	ppath := path        // Pristine path

	for i, lcpIndex := 0, len(path); i < lcpIndex; i++ {
		if path[i] == ':' {
			if i > 0 && path[i-1] == '\\' {
				path = path[:i-1] + path[i:]
				i--
				lcpIndex--
				continue
			}
			j := i + 1

			r.insert(method, path[:i], nil, staticKind, "", nil)
			for ; i < lcpIndex && path[i] != '/'; i++ {
			}

			pnames = append(pnames, path[j:i])
			path = path[:j] + path[i:]
			i, lcpIndex = j, len(path)

			if i == lcpIndex {
				// path node is last fragment of route path. ie. `/users/:id`
				r.insert(method, path[:i], h, paramKind, ppath, pnames)
			} else {
				r.insert(method, path[:i], nil, paramKind, "", nil)
			}
		} else if path[i] == '*' {
			r.insert(method, path[:i], nil, staticKind, "", nil)
			pnames = append(pnames, "*")
			r.insert(method, path[:i+1], h, anyKind, ppath, pnames)
		}
	}

	r.insert(method, path, h, staticKind, ppath, pnames)
}

func (r *router) insert(method, path string, h HandlerFunc, t kind, ppath string, pnames []string) {
	// Copied from github.com/labstack/echo's Router.insert method
	// Adjust max param
	if paramLen := len(pnames); *r.broker.maxParam < paramLen {
		*r.broker.maxParam = paramLen
	}

	currentNode := r.tree // Current node as root
	if currentNode == nil {
		panic("echo: invalid method")
	}
	search := path

	for {
		searchLen := len(search)
		prefixLen := len(currentNode.prefix)
		lcpLen := 0

		// LCP - Longest Common Prefix (https://en.wikipedia.org/wiki/LCP_array)
		max := prefixLen
		if searchLen < max {
			max = searchLen
		}
		for ; lcpLen < max && search[lcpLen] == currentNode.prefix[lcpLen]; lcpLen++ {
		}

		if lcpLen == 0 {
			// At root node
			currentNode.label = search[0]
			currentNode.prefix = search
			if h != nil {
				currentNode.kind = t
				currentNode.addHandler(method, h)
				currentNode.ppath = ppath
				currentNode.pnames = pnames
			}
			currentNode.isLeaf = currentNode.isLeafNode()
		} else if lcpLen < prefixLen {
			// Split node
			n := newNode(
				currentNode.kind,
				currentNode.prefix[lcpLen:],
				currentNode,
				currentNode.staticChildren,
				currentNode.methodHandler,
				currentNode.ppath,
				currentNode.pnames,
				currentNode.paramChild,
				currentNode.anyChild,
			)
			// Update parent path for all children to new node
			for _, child := range currentNode.staticChildren {
				child.parent = n
			}
			if currentNode.paramChild != nil {
				currentNode.paramChild.parent = n
			}
			if currentNode.anyChild != nil {
				currentNode.anyChild.parent = n
			}

			// Reset parent node
			currentNode.kind = staticKind
			currentNode.label = currentNode.prefix[0]
			currentNode.prefix = currentNode.prefix[:lcpLen]
			currentNode.staticChildren = nil
			currentNode.methodHandler = new(methodHandler)
			currentNode.ppath = ""
			currentNode.pnames = nil
			currentNode.paramChild = nil
			currentNode.anyChild = nil
			currentNode.isLeaf = false
			currentNode.isHandler = false

			// Only Static children could reach here
			currentNode.addStaticChild(n)

			if lcpLen == searchLen {
				// At parent node
				currentNode.kind = t
				currentNode.addHandler(method, h)
				currentNode.ppath = ppath
				currentNode.pnames = pnames
			} else {
				// Create child node
				n = newNode(t, search[lcpLen:], currentNode, nil, new(methodHandler), ppath, pnames, nil, nil)
				n.addHandler(method, h)
				// Only Static children could reach here
				currentNode.addStaticChild(n)
			}
			currentNode.isLeaf = currentNode.isLeafNode()
		} else if lcpLen < searchLen {
			search = search[lcpLen:]
			c := currentNode.findChildWithLabel(search[0])
			if c != nil {
				// Go deeper
				currentNode = c
				continue
			}
			// Create child node
			n := newNode(t, search, currentNode, nil, new(methodHandler), ppath, pnames, nil, nil)
			n.addHandler(method, h)
			switch t {
			case staticKind:
				currentNode.addStaticChild(n)
			case paramKind:
				currentNode.paramChild = n
			case anyKind:
				currentNode.anyChild = n
			}
			currentNode.isLeaf = currentNode.isLeafNode()
		} else {
			// Node already exists
			if h != nil {
				currentNode.addHandler(method, h)
				currentNode.ppath = ppath
				if len(currentNode.pnames) == 0 { // Issue #729
					currentNode.pnames = pnames
				}
			}
		}
		return
	}
}

func (r *router) Find(method, path string, ctx *context) {
	// Copied from github.com/labstack/echo's Router.Find method
	ctx.path = path
	currentNode := r.tree // Current node as root

	var (
		previousBestMatchNode *node
		matchedHandler        HandlerFunc
		// search stores the remaining path to check for match. By each iteration we move from start of
		// path to end of the path and search value gets shorter and shorter.
		search      = path
		searchIndex = 0
		paramIndex  int           // Param counter
		paramValues = ctx.pvalues // Use internal slice to give illusion of a dynamic slice in interface
	)

	// Backtracking is needed when a dead end (leaf node) is reached in the router tree.
	// To backtrack the current node will be changed to the parent node and the next kind for the
	// router logic will be returned based on fromKind or kind of the dead end node
	// (static > param > any).  For example if there is no static node match we should check parent
	// next sibling by kind (param).  Backtracking itself does not check if there is a next sibling,
	// this is done by the router logic.
	backtrackToNextNodeKind := func(fromKind kind) (nextNodeKind kind, valid bool) {
		previous := currentNode
		currentNode = previous.parent
		valid = currentNode != nil

		// Next node type by priority
		if previous.kind == anyKind {
			nextNodeKind = staticKind
		} else {
			nextNodeKind = previous.kind + 1
		}

		if fromKind == staticKind {
			// when backtracking is done from static kind block we did not change search so nothing to
			// restore
			return
		}

		// restore search to value it was before we move to current node we are backtracking from.
		if previous.kind == staticKind {
			searchIndex -= len(previous.prefix)
		} else {
			paramIndex--
			// for param/any node.prefix value is always `:` so we can not deduce searchIndex from that
			// and must use pValue for that index as it would also contain part of path we cut off before
			// moving into node we are backtracking from
			searchIndex -= len(paramValues[paramIndex])
			paramValues[paramIndex] = ""
		}
		search = path[searchIndex:]
		return
	}

	// Router tree is implemented by longest common prefix array (LCP array)
	// https://en.wikipedia.org/wiki/LCP_array
	// Tree search is implemented as for loop where one loop iteration is divided into 3 separate
	// blocks
	// Each of these blocks checks specific kind of node (static/param/any). Order of blocks reflex
	// their priority in routing.
	// Search order/priority is: static > param > any.
	//
	// Note: backtracking in tree is implemented by replacing/switching currentNode to previous node
	// and hoping to (goto statement) next block by priority to check if it is the match.
	for {
		prefixLen := 0 // Prefix length
		lcpLen := 0    // LCP (longest common prefix) length

		if currentNode.kind == staticKind {
			searchLen := len(search)
			prefixLen = len(currentNode.prefix)

			// LCP - Longest Common Prefix (https://en.wikipedia.org/wiki/LCP_array)
			max := prefixLen
			if searchLen < max {
				max = searchLen
			}
			for ; lcpLen < max && search[lcpLen] == currentNode.prefix[lcpLen]; lcpLen++ {
			}
		}

		if lcpLen != prefixLen {
			// No matching prefix, let's backtrack to the first possible alternative node of the decision
			// path
			nk, ok := backtrackToNextNodeKind(staticKind)
			if !ok {
				return // No other possibilities on the decision path
			} else if nk == paramKind {
				goto Param
				// NOTE: this case (backtracking from static node to previous any node) can not happen by
				// current any matching logic. Any node is end of search currently
				//} else if nk == anyKind {
				//	goto Any
			} else {
				// Not found (this should never be possible for static node we are looking currently)
				break
			}
		}

		// The full prefix has matched, remove the prefix from the remaining search
		search = search[lcpLen:]
		searchIndex = searchIndex + lcpLen

		// Finish routing if no remaining search and we are on a node with handler and matching method
		// type
		if search == "" && currentNode.isHandler {
			// check if current node has handler registered for http method we are looking for. we store
			// currentNode as best matching in case we do no find no more routes matching this path+method
			if previousBestMatchNode == nil {
				previousBestMatchNode = currentNode
			}
			if h := currentNode.findHandler(method); h != nil {
				matchedHandler = h
				break
			}
		}

		// Static node
		if search != "" {
			if child := currentNode.findStaticChild(search[0]); child != nil {
				currentNode = child
				continue
			}
		}

	Param:
		// Param node
		if child := currentNode.paramChild; search != "" && child != nil {
			currentNode = child
			i := 0
			l := len(search)
			if currentNode.isLeaf {
				// when param node does not have any children then param node should act similarly to any
				// node - consider all remaining search as match
				i = l
			} else {
				for ; i < l && search[i] != '/'; i++ {
				}
			}

			paramValues[paramIndex] = search[:i]
			paramIndex++
			search = search[i:]
			searchIndex = searchIndex + i
			continue
		}

	Any:
		// Any node
		if child := currentNode.anyChild; child != nil {
			// If any node is found, use remaining path for paramValues
			currentNode = child
			paramValues[len(currentNode.pnames)-1] = search
			// update indexes/search in case we need to backtrack when no handler match is found
			paramIndex++
			searchIndex += +len(search)
			search = ""

			// check if current node has handler registered for http method we are looking for. we store
			// currentNode as best matching in case we do no find no more routes matching this path+method
			if previousBestMatchNode == nil {
				previousBestMatchNode = currentNode
			}
			if h := currentNode.findHandler(method); h != nil {
				matchedHandler = h
				break
			}
		}

		// Let's backtrack to the first possible alternative node of the decision path
		nk, ok := backtrackToNextNodeKind(anyKind)
		if !ok {
			break // No other possibilities on the decision path
		} else if nk == paramKind {
			goto Param
		} else if nk == anyKind {
			goto Any
		} else {
			// Not found
			break
		}
	}

	if currentNode == nil && previousBestMatchNode == nil {
		return // nothing matched at all
	}

	if matchedHandler != nil {
		ctx.handler = matchedHandler
	} else {
		// use previous match as basis. although we have no matching handler we have path match.
		// so we can send http.StatusMethodNotAllowed (405) instead of http.StatusNotFound (404)
		currentNode = previousBestMatchNode

		ctx.handler = NotFoundHandler
	}
	ctx.path = currentNode.ppath
	ctx.pnames = currentNode.pnames
}
