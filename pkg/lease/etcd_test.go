// Copyright 2014 The fleet Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lease

import (
	"reflect"
	"testing"
	"time"

	etcd "github.com/coreos/etcd/client"
)

func TestSerializeLeaseMetadata(t *testing.T) {
	tests := []struct {
		machID string
		ver    int
		want   string
	}{
		{
			machID: "XXX",
			ver:    9,
			want:   `{"MachineID":"XXX","Version":9}`,
		},
		{
			machID: "XXX",
			ver:    0,
			want:   `{"MachineID":"XXX","Version":0}`,
		},
	}

	for i, tt := range tests {
		got, err := serializeLeaseMetadata(tt.machID, tt.ver)
		if err != nil {
			t.Errorf("case %d: unexpected err=%v", i, err)
			continue
		}
		if tt.want != got {
			t.Errorf("case %d: incorrect output from serializeLeaseMetadata\nwant=%s\ngot=%s", i, tt.want, got)
		}
	}
}

func TestLeaseFromResponse(t *testing.T) {
	tests := []struct {
		res  etcd.Response
		want etcdLease
	}{
		// typical case
		{
			res: etcd.Response{
				Node: &etcd.Node{
					Key:           "/foo/bar",
					ModifiedIndex: 12,
					TTL:           9,
					Value:         `{"MachineID":"XXX","Version":19}`,
				},
			},
			want: etcdLease{
				key: "/foo/bar",
				idx: 12,
				ttl: time.Second * 9,
				meta: etcdLeaseMetadata{
					MachineID: "XXX",
					Version:   19,
				},
			},
		},

		// backwards-compatibility with unversioned engines
		{
			res: etcd.Response{
				Node: &etcd.Node{
					Key:           "/foo/bar",
					ModifiedIndex: 12,
					TTL:           9,
					Value:         "XXX",
				},
			},
			want: etcdLease{
				key: "/foo/bar",
				idx: 12,
				ttl: time.Second * 9,
				meta: etcdLeaseMetadata{
					MachineID: "XXX",
					Version:   0,
				},
			},
		},

		// json decode failures are treated like a nonversioned lease
		{
			res: etcd.Response{
				Node: &etcd.Node{
					Key:           "/foo/bar",
					ModifiedIndex: 12,
					TTL:           9,
					Value:         `{"MachineID":"XXX","Ver`,
				},
			},
			want: etcdLease{
				key: "/foo/bar",
				idx: 12,
				ttl: time.Second * 9,
				meta: etcdLeaseMetadata{
					MachineID: `{"MachineID":"XXX","Ver`,
					Version:   0,
				},
			},
		},
	}

	for i, tt := range tests {
		var r *etcdLeaseManager
		got := r.leaseFromResponse(&tt.res)
		if !reflect.DeepEqual(tt.want, *got) {
			t.Errorf("case %d: incorrect output from leaseFromResponse\nwant=%#v\ngot=%#vs", i, tt.want, *got)
		}
	}
}
