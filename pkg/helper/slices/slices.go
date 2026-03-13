package slices

import "sync"

// Equal reports whether two slices are equal: the same length and all
// elements equal. If the lengths are different, Equal returns false.
// Otherwise, the elements are compared in increasing index order, and the
// comparison stops at the first unequal pair.
// Floating point NaNs are not considered equal.
func Equal[E comparable](s1, s2 []E) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}

// EqualFunc reports whether two slices are equal using a comparison
// function on each pair of elements. If the lengths are different,
// EqualFunc returns false. Otherwise, the elements are compared in
// increasing index order, and the comparison stops at the first index
// for which eq returns false.
func EqualFunc[E1, E2 any](s1 []E1, s2 []E2, eq func(E1, E2) bool) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i, v1 := range s1 {
		v2 := s2[i]
		if !eq(v1, v2) {
			return false
		}
	}
	return true
}

// Index returns the index of the first occurrence of v in s,
// or -1 if not present.
func Index[E comparable](s []E, v E) int {
	for i := range s {
		if v == s[i] {
			return i
		}
	}
	return -1
}

// IndexFunc returns the first index i satisfying f(s[i]),
// or -1 if none do.
func IndexFunc[E any](s []E, f func(E) bool) int {
	for i := range s {
		if f(s[i]) {
			return i
		}
	}
	return -1
}

// Contains reports whether v is present in s.
func Contains[E comparable](s []E, v E) bool {
	return Index(s, v) >= 0
}

// ContainsFunc reports whether at least one
// element e of s satisfies f(e).
func ContainsFunc[E any](s []E, f func(E) bool) bool {
	return IndexFunc(s, f) >= 0
}

// Delete removes the elements s[i:j] from s, returning the modified slice.
// Delete panics if s[i:j] is not a valid slice of s.
// Delete modifies the contents of the slice s; it does not create a new slice.
// Delete is O(len(s)-j), so if many items must be deleted, it is better to
// make a single call deleting them all together than to delete one at a time.
// Delete might not modify the elements s[len(s)-(j-i):len(s)]. If those
// elements contain pointers you might consider zeroing those elements so that
// objects they reference can be garbage collected.
func Delete[S ~[]E, E any](s S, i, j int) S {
	_ = s[i:j] // bounds check
	return append(s[:i], s[j:]...)
}

// Replace replaces the elements s[i:j] by the given v, and returns the
// modified slice. Replace panics if s[i:j] is not a valid slice of s.
func Replace[S ~[]E, E any](s S, i, j int, v ...E) S {
	_ = s[i:j] // verify that i:j is a valid subslice
	tot := len(s[:i]) + len(v) + len(s[j:])
	if tot <= cap(s) {
		s2 := s[:tot]
		copy(s2[i+len(v):], s[j:])
		copy(s2[i:], v)
		return s2
	}
	s2 := make(S, tot)
	copy(s2, s[:i])
	copy(s2[i:], v)
	copy(s2[i+len(v):], s[j:])
	return s2
}

// Clone returns a copy of the slice.
// The elements are copied using assignment, so this is a shallow clone.
func Clone[S ~[]E, E any](s S) S {
	// Preserve nil in case it matters.
	if s == nil {
		return nil
	}
	return append([]E{}, s...)
}

// Clip removes unused capacity from the slice, returning s[:len(s):len(s)].
func Clip[S ~[]E, E any](s S) S {
	return s[:len(s):len(s)]
}

// Intersect 查交集,会去重
func Intersect[S ~[]E, E comparable](s1, s2 S) S {
	l1 := len(s1)
	l2 := len(s2)
	if l1 == 0 || l2 == 0 {
		return make(S, 0)
	}
	if l1 > l2 {
		s1, s2 = s2, s1
		l1, l2 = l2, l1
	}
	ret := make(S, 0, l2)
	mp := make(map[E]struct{}, l1)
	for i := range s1 {
		if _, ok := mp[s1[i]]; !ok {
			mp[s1[i]] = struct{}{}
		}
	}
	uMp := make(map[E]struct{}, len(mp))
	for i := range s2 {
		if _, ok := mp[s2[i]]; ok {
			//排重
			if _, _ok := uMp[s2[i]]; !_ok {
				uMp[s2[i]] = struct{}{}
				ret = append(ret, s2[i])
			}
		}
	}
	return ret
}

// Union 并集,会去重
func Union[S ~[]E, E comparable](s1, s2 S) S {
	ret := make(S, 0, len(s1)+len(s2))
	mp := make(map[E]struct{})
	for i := range s1 {
		if _, ok := mp[s1[i]]; !ok {
			mp[s1[i]] = struct{}{}
			ret = append(ret, s1[i])
		}
	}
	for i := range s2 {
		if _, ok := mp[s2[i]]; !ok {
			mp[s2[i]] = struct{}{}
			ret = append(ret, s2[i])
		}
	}
	return Clip(ret)
}

// Difference 差集,会去重
func Difference[S ~[]E, E comparable](s1, s2 S) S {
	if len(s1) > len(s2) {
		s1, s2 = s2, s1
	}
	ret := make(S, 0, len(s1)+len(s2))
	mp := make(map[E]struct{})
	for i := range s1 {
		if _, ok := mp[s1[i]]; !ok {
			mp[s1[i]] = struct{}{}
			ret = append(ret, s1[i])
		}
	}
	diffMp := make(map[E]struct{})
	for i := range s2 {
		if _, ok := diffMp[s2[i]]; ok {
			continue
		}
		diffMp[s2[i]] = struct{}{}
		if _, ok := mp[s2[i]]; ok {
			idx := Index(ret, s2[i])
			ret = append(ret[:idx], ret[idx+1:]...)
			continue
		}
		ret = append(ret, s2[i])
	}
	return Clip(ret)
}

// Map 遍历处理
func Map[S ~[]E, E any, T any](s S, f func(v E, k int) T) []T {
	res := make([]T, len(s))
	for k, v := range s {
		res[k] = f(v, k)
	}
	return res
}

// MapAsync 异步遍历处理
func MapAsync[S ~[]E, E any, T any](s S, f func(v E, k int) T) []T {
	res := make([]T, len(s))
	mch := make(chan struct {
		k int
		v T
	})
	for k, v := range s {
		go func(k int, v E) {
			mch <- struct {
				k int
				v T
			}{
				k: k,
				v: f(v, k),
			}
		}(k, v)
	}
	for i := 0; i < len(s); i++ {
		v := <-mch
		res[v.k] = v.v
	}
	return res
}

// MapAsyncN 异步遍历处理,同时控制最大并发数
func MapAsyncN[S ~[]E, E any, T any](s S, f func(v E, k int) T, n uint) []T {
	if n == 0 {
		return MapAsync(s, f)
	}
	res := make([]T, len(s))
	mch := make(chan struct {
		k int
		v T
	})
	lch := make(chan struct{}, n)
	go func() {
		for i := 0; i < len(s); i++ {
			v := <-mch
			res[v.k] = v.v
		}
	}()
	wg := sync.WaitGroup{}
	wg.Add(len(s))
	for k, v := range s {
		lch <- struct{}{}
		go func(k int, v E) {
			defer func() {
				<-lch
				wg.Done()
			}()
			mch <- struct {
				k int
				v T
			}{
				k: k,
				v: f(v, k),
			}
		}(k, v)
	}
	wg.Wait()
	return res
}

func Foreach[S ~[]E, E any](s S, f func(e E, k int) error) error {
	for i, v := range s {
		if err := f(v, i); err != nil {
			return err
		}
	}
	return nil
}

func Reduce[E any, T any](s []E, f func(d *T, v E, k int)) T {
	res := new(T)
	for k, v := range s {
		f(res, v, k)
	}
	return *res
}

// Unique 切片过滤重复
func Unique[S ~[]E, E comparable](s S) S {
	tmp := make(map[E]struct{}, 0)
	r := make(S, 0, len(s))
	for i := range s {
		l := len(tmp)
		tmp[s[i]] = struct{}{}
		if l != len(tmp) {
			r = append(r, s[i])
		}
	}
	return Clip(r)
}
