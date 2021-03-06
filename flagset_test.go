// Copyright 2021 Marko Kungla. All rights reserved.
// Use of this source code is governed by a The Apache-style
// license that can be found in the LICENSE file.

package varflag

import (
	"os"
	"testing"
)

//nolint: funlen, cyclop
func TestFlagSet(t *testing.T) {
	args := []string{
		os.Args[0], "cmd1", "--flag1", "val1", "--flag2", "flag2-value",
		"arg1", "--flag3=on", "-v", // global flag can be any place
		"subcmd", "--flag4", "val 4 flag", "arg2", "arg3", "-x", "on", // global flag can be any place
	}

	global, err := NewFlagSet(args[0], 0)
	if err != nil {
		t.Error("did not expect error got ", err)
	}
	v, _ := Bool("verbose", false, "increase verbosity", "v")
	x, _ := Bool("x", false, "print commands")
	r, _ := Bool("random", false, "random flag")
	global.Add(v, x, r)

	flag1, _ := New("flag1", "", "first flag for first cmd")
	flag2, _ := New("flag2", "", "another flag for first cmd")
	flag3, _ := Bool("flag3", false, "bool flag for first command")
	cmd1, err := NewFlagSet("cmd1", 1)
	if err != nil {
		t.Error("did not expect error got ", err)
	}
	cmd1.Add(flag1, flag2, flag3)

	flag5, _ := New("flag5", "", "flag5 for second cmd")
	cmd2, err := NewFlagSet("cmd2", 0)
	if err != nil {
		t.Error("did not expect error got ", err)
	}
	cmd2.Add(flag5)

	subcmd, err := NewFlagSet("subcmd", 1)
	if err != nil {
		t.Error("did not expect error got ", err)
	}
	flag4, _ := New("flag4", "", "flag4 for sub command")
	subcmd.Add(flag4)
	cmd1.AddSet(subcmd)
	global.AddSet(cmd1, cmd2)

	if err := global.Parse(args); err != nil {
		t.Error("did not expect error got ", err)
	}
	if !v.Present() || !v.Value() {
		t.Error("expected verbose flag", v.Present(), v.Value())
	}
	if !x.Present() || !x.Value() {
		t.Error("expected x flag", x.Present(), x.Value())
	}
	if r.Present() || r.Value() {
		t.Error("expected no r flag", r.Present(), r.Value())
	}
	if !cmd1.Present() {
		t.Error("expected cmd1 to be present", cmd1.Present())
	}
	if !subcmd.Present() {
		t.Error("expected subcmd to be present", subcmd.Present())
	}
	if cmd2.Present() {
		t.Error("expected cmd2 not to be present", cmd2.Present())
	}
	if cmd1.Name() != "cmd1" {
		t.Error("expected cmd name cmd1 got ", cmd1.Name())
	}
	if subcmd.Name() != "subcmd" {
		t.Error("expected subcmd name subcmd got ", subcmd.Name())
	}

	if !flag1.Present() {
		t.Error("expected flag1 ", flag1.Present(), flag1.Value())
	}
	if !flag2.Present() {
		t.Error("expected flag2 ", flag2.Present(), flag2.Value())
	}
	if !flag3.Present() {
		t.Error("expected flag2 ", flag3.Present(), flag3.Value())
	}
	if !flag4.Present() {
		t.Error("expected flag4 ", flag4.Present(), flag4.Value())
	}
	if flag5.Present() {
		t.Error("expected no flag5", flag4.Present(), flag4.Value())
	}

	if len(global.Args()) != 0 {
		t.Error("expected no global args got ", global.Args())
	}
	if len(cmd2.Args()) != 0 {
		t.Error("expected no cmd2 args got ", cmd2.Args())
	}
	if len(cmd1.Args()) != 1 {
		t.Error("expected cmd1 to have 1 arg got ", cmd1.Args())
	}
	if len(subcmd.Args()) != 2 {
		t.Error("expected subcmd to have 2 arg got ", subcmd.Args())
	}
	if global.GetActiveSetName() != "subcmd" {
		t.Error("expected active cmd/set to be subcmd got ", global.GetActiveSetName())
	}
	if subcmd.Pos() != 2 {
		t.Error("expected subcmd pos to be 2 got ", subcmd.Pos())
	}
}

func TestFlagSetName(t *testing.T) {
	for _, tt := range testflags() {
		t.Run(tt.name, func(t *testing.T) {
			flag, err := NewFlagSet(tt.name, 0)
			if tt.valid {
				if err != nil {
					t.Errorf("valid flag set name %q did not expect error got %q", tt.name, err)
				}
				if n := flag.Name(); n != tt.name {
					t.Errorf("flag set name should be %q got %q", tt.name, n)
				}
				return
			}
			if err == nil {
				t.Errorf("invalid flag set %q expected error got <nil>", tt.name)
			}
			if flag != nil {
				t.Errorf("invalid flag set %q should be <nil> got %#v", tt.name, flag)
			}
		})
	}
}

func TestShaddowFlags(t *testing.T) {
	args := []string{
		os.Args[0], "cmd2", "--flag", "val2",
	}

	global, err := NewFlagSet(args[0], 0)
	if err != nil {
		t.Error("did not expect error got ", err)
	}
	cmd1, err := NewFlagSet("cmd1", 1)
	if err != nil {
		t.Error("did not expect error got ", err)
	}
	cmd2, err := NewFlagSet("cmd2", 1)
	if err != nil {
		t.Error("did not expect error got ", err)
	}

	f1, _ := New("flag", "", "")
	cmd1.Add(f1)
	f2, _ := New("flag", "", "")
	cmd2.Add(f2)

	global.AddSet(cmd1, cmd2)
	if err := global.Parse(args); err != nil {
		t.Error("did not expect error got ", err)
	}

	if cmd1.Name() != "cmd1" {
		t.Error("expected cmd name cmd1 got ", cmd1.Name())
	}

	if cmd2.Name() != "cmd2" {
		t.Error("expected cmd name cmd2 got ", cmd2.Name())
	}

	if cmd1.Present() {
		t.Error("expected cmd1 not to be present", cmd1.Present())
	}

	if !cmd2.Present() {
		t.Error("expected cmd2 to be present", cmd2.Present())
	}

	if !f2.Present() {
		t.Error("expected cmd2 flag present: ", f2.Present(), f2.Value())
	}

	if f1.Present() {
		t.Error("did not expect cmd1 flag to be present :", f1.Present(), f1.Value())
	}

	if f1.Value() != "" {
		t.Error("f1 value should be empty got", f1.Value())
	}
	if f2.Value() != "val2" {
		t.Error("f2 value should val2 got", f2.Value())
	}
}
