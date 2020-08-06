package http

import (
	"errors"
	"strings"
)

type TrieMatchingMode int

const (
	EXPLICIT_NODES_ONLY TrieMatchingMode = iota
	WILDCARD_ROOT_NODES_ALLOWED
	WILDCARD_LEAF_NODES_ALLOWED
	WILDCARD_NODES_ALLOWED
)

func isNamedWildCard(key string) bool {
	return strings.IndexRune(key, '{') != -1 && strings.IndexRune(key, '}') != -1
}

const (
	SEPARATOR = "/"
	WILDCARD  = "*"
)

type trieNode struct {
	key           string
	value         interface{}
	wildcard      string
	children      map[string]*trieNode
	namedWildcard string
}

func newTrieNode(key string, value interface{}, wildcard string) *trieNode {
	var namedWildcard string
	if isNamedWildCard(key) {
		namedWildcard = key[strings.IndexRune(key, '{')+1 : strings.IndexRune(key, '}')]
	}
	return &trieNode{
		key:           key,
		wildcard:      wildcard,
		value:         value,
		children:      map[string]*trieNode{},
		namedWildcard: namedWildcard,
	}
}

func (n *trieNode) addInnerChild(key string, child *trieNode) {
	n.children[key] = child
}

func (n *trieNode) updateKeyWithNamedWildcard(key string) {
	n.key = key
	newNamedWildcard := key[strings.IndexRune(key, '{')+1 : strings.IndexRune(key, '}')]
	n.namedWildcard = newNamedWildcard
}

func (n *trieNode) insert(path []string, index int, value interface{}) {
	if index == len(path) {
		return
	}

	token := path[index]
	key := token
	if isNamedWildCard(token) {
		key = n.wildcard
	}
	node, exists := n.children[key]
	if !exists {
		var nodeValue interface{} = nil
		if index == len(path)-1 {
			nodeValue = value
		}
		node = newTrieNode(token, nodeValue, n.wildcard)
		n.addInnerChild(key, node)
	} else {
		if isNamedWildCard(token) {
			node.updateKeyWithNamedWildcard(token)
		}

		if index == len(path)-1 {
			node.value = value
		}
	}

	node.insert(path, index+1, value)
}

func (n *trieNode) insertOrUpdate() {

}

func (n *trieNode) retrieve(path []string, index int, params map[string]string, mode TrieMatchingMode) interface{} {
	if index >= len(path) {
		return nil
	}

	token := path[index]
	node, exists := n.children[token]

	usedWildcard := false
	if !exists {
		if mode == WILDCARD_NODES_ALLOWED {
			node, exists = n.children[n.wildcard]
			if !exists {
				return nil
			}
			usedWildcard = true
		} else if mode == WILDCARD_ROOT_NODES_ALLOWED && index == 1 {
			node, exists = n.children[n.wildcard]
			if !exists {
				return nil
			}
			usedWildcard = true
		} else if mode == WILDCARD_LEAF_NODES_ALLOWED && index+1 == len(path) {
			node, exists = n.children[n.wildcard]
			if !exists {
				return nil
			}
			usedWildcard = true
		} else {
			return nil
		}
	} else {
		if _, exists = n.children[n.wildcard]; index+1 == len(path) && node.value == nil && exists &&
			mode != EXPLICIT_NODES_ONLY && mode != WILDCARD_ROOT_NODES_ALLOWED {
			node = n.children[n.wildcard]
			usedWildcard = true
		} else if _, exists = n.children[n.wildcard]; index == 1 && node.value == nil && exists &&
			mode == WILDCARD_ROOT_NODES_ALLOWED {
			node = n.children[n.wildcard]
			usedWildcard = true
		} else {
			usedWildcard = token == n.wildcard
		}
	}
	n.put(params, node, token)
	if index == len(path)-1 {
		return node.value
	}
	nodeValue := node.retrieve(path, index+1, params, mode)
	if nodeValue == nil && !usedWildcard && mode != EXPLICIT_NODES_ONLY {
		node, exists = n.children[n.wildcard]
		if exists {
			n.put(params, node, token)
			nodeValue = node.retrieve(path, index+1, params, mode)
		}
	}

	return nodeValue
}

func (n *trieNode) put(params map[string]string, node *trieNode, value string) {
	if params != nil && node.namedWildcard != "" {
		params[node.namedWildcard] = value
	}
}

type pathTrie struct {
	root      *trieNode
	rootValue interface{}
}

func newPathTrie() *pathTrie {
	return &pathTrie{
		root: newTrieNode(SEPARATOR, nil, WILDCARD),
	}
}

func (t *pathTrie) insert(path string, value interface{}) {
	if path == "" || path == "/" {
		t.rootValue = value
		return
	}
	strs := strings.Split(path, SEPARATOR)
	index := 0
	if strs[0] == "" {
		index = 1
	}
	t.root.insert(strs, index, value)
}

func (t *pathTrie) insertOrUpdate() {

}

func (t *pathTrie) retrieve(path string, mode TrieMatchingMode) (interface{}, map[string]string) {
	params := map[string]string{}
	if path == "" || path == "/" {
		return t.rootValue, params
	}
	strs := strings.Split(path, SEPARATOR)
	index := 0
	if strs[0] == "" {
		index = 1
	}

	return t.root.retrieve(strs, index, params, mode), params
}

func (t *pathTrie) retrieveAll(path string) func() (interface{}, map[string]string, error) {
	mode := EXPLICIT_NODES_ONLY
	return func() (interface{}, map[string]string, error) {
		if mode > WILDCARD_NODES_ALLOWED {
			return nil, nil, errors.New("NoRoute")
		}
		handler, params := t.retrieve(path, mode)
		mode += 1
		return handler, params, nil
	}
}
