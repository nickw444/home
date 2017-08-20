package characteristic

import (
	"net"
	"testing"
)

func TestCharacteristicUpdateValuesOfWrongType(t *testing.T) {
	c := NewCharacteristic(TypeOn)
	c.Value = 5

	c.UpdateValue(float64(20.5))

	if is, want := c.Value, 20; is != want {
		t.Fatalf("is=%v want=%v", is, want)
	}

	c.UpdateValue("91")

	if is, want := c.Value, 91; is != want {
		t.Fatalf("is=%v want=%v", is, want)
	}

	c.UpdateValue(true)

	if is, want := c.Value, 1; is != want {
		t.Fatalf("is=%v want=%v", is, want)
	}
}

func TestCharacteristicLocalDelegate(t *testing.T) {
	c := NewCharacteristic(TypeOn)
	c.Value = 5

	var oldValue interface{}
	var newValue interface{}

	c.OnValueUpdate(func(c *Characteristic, new, old interface{}) {
		newValue = new
		oldValue = old
	})

	c.UpdateValue(10)
	c.UpdateValueFromConnection(20, TestConn)

	if is, want := oldValue, 5; is != want {
		t.Fatalf("is=%v want=%v", is, want)
	}
	if is, want := newValue, 10; is != want {
		t.Fatalf("is=%v want=%v", is, want)
	}
}

func TestCharacteristicRemoteDelegate(t *testing.T) {
	c := NewCharacteristic(TypeOn)
	c.Perms = PermsAll()
	c.Value = 5

	var oldValue interface{}
	var newValue interface{}
	c.OnValueUpdateFromConn(func(conn net.Conn, c *Characteristic, new, old interface{}) {
		if conn != TestConn {
			t.Fatal(conn)
		}
		newValue = new
		oldValue = old
	})

	c.UpdateValueFromConnection(10, TestConn)
	c.UpdateValue(20)

	if is, want := oldValue, 5; is != want {
		t.Fatalf("is=%v want=%v", is, want)
	}
	if is, want := newValue, 10; is != want {
		t.Fatalf("is=%v want=%v", is, want)
	}
}

func TestNoValueChange(t *testing.T) {
	c := NewCharacteristic(TypeOn)
	c.Value = 5

	changed := false
	c.OnValueUpdateFromConn(func(conn net.Conn, c *Characteristic, new, old interface{}) {
		if conn != TestConn {
			t.Fatal(conn)
		}
		changed = true
	})

	c.OnValueUpdate(func(c *Characteristic, new, old interface{}) {
		changed = true
	})

	c.UpdateValue(5)
	c.UpdateValueFromConnection(5, TestConn)

	if changed != false {
		t.Fatal(changed)
	}
}

func TestReadOnlyValue(t *testing.T) {
	c := NewCharacteristic(TypeOn)
	c.Perms = PermsRead()
	c.Value = 5

	remoteChanged := false
	localChanged := false
	c.OnValueUpdateFromConn(func(conn net.Conn, c *Characteristic, new, old interface{}) {
		if conn != TestConn {
			t.Fatal(conn)
		}
		remoteChanged = true
	})

	c.OnValueUpdate(func(c *Characteristic, new, old interface{}) {
		localChanged = true
	})

	c.UpdateValue(10)
	c.UpdateValueFromConnection(11, TestConn)

	if is, want := c.Value, 10; is != want {
		t.Fatalf("is=%v want=%v", is, want)
	}
	if remoteChanged != false {
		t.Fatal(remoteChanged)
	}
	if localChanged != true {
		t.Fatal(localChanged)
	}
}

func TestEqual(t *testing.T) {
	c1 := NewCharacteristic(TypeOn)
	c1.Value = 5

	c2 := NewCharacteristic(TypeOn)
	c2.Value = 5

	if c1.Equal(c2) == false {
		t.Fatal("characteristics not the same")
	}
}
