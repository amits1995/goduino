package goduino

import (
	"fmt"
	"github.com/argandas/goduino/firmata"
	"github.com/tarm/serial"
	"io"
	"log"
	"os"
	"time"
)

const (
	Input  = firmata.Input
	Output = firmata.Output
	Analog = firmata.Analog
	Pwm    = firmata.Pwm
	Servo  = firmata.Servo
)

type firmataBoard interface {
	Connect(io.ReadWriteCloser) error
	Disconnect() error
	Pins() []firmata.Pin
	AnalogWrite(int, int) error
	SetPinMode(int, int) error
	ReportAnalog(int, int) error
	ReportDigital(int, int) error
	DigitalWrite(int, int) error
	I2cRead(int, int) error
	I2cWrite(int, []byte) error
	I2cConfig(int) error
	ServoConfig(int, int, int) error
}

// Arduino Firmata client for golang
type Goduino struct {
	name    string
	port    string
	board   firmataBoard
	conn    io.ReadWriteCloser
	openSP  func(port string) (io.ReadWriteCloser, error)
	logger  *log.Logger
	verbose bool
}

// Creates a new Goduino object and connects to the Arduino board
// over specified serial port. This function blocks till a connection is
// succesfullt established and pin mappings are retrieved.
func New(name string, args ...interface{}) *Goduino {
	// Create new Goduino client
	goduino := &Goduino{
		name:  name,
		port:  "",
		conn:  nil,
		board: firmata.New(),
		openSP: func(port string) (io.ReadWriteCloser, error) {
			return serial.OpenPort(&serial.Config{Name: port, Baud: 57600})
		},
		logger:  log.New(os.Stdout, fmt.Sprintf("[%s] ", name), log.Ltime),
		verbose: true,
	}
	// Parse variadic args
	for _, arg := range args {
		switch arg.(type) {
		case string:
			goduino.port = arg.(string)
		case io.ReadWriteCloser:
			goduino.conn = arg.(io.ReadWriteCloser)
		}
	}
	return goduino
}

// Connect starts a connection to the firmata board.
func (ino *Goduino) Connect() error {
	if ino.conn == nil {
		// Try to connect to serial port
		sp, err := ino.openSP(ino.Port())
		if err != nil {
			return err
		}
		// Serial connection was successful
		ino.conn = sp
	}
	// Firmata connection
	return ino.board.Connect(ino.conn)
}

// Disconnect closes the io connection to the firmata board
func (ino *Goduino) Disconnect() (err error) {
	if ino.board != nil {
		// Disconnect firmata board
		return ino.board.Disconnect()
	}
	return nil
}

// Port returns the  FirmataAdaptors port
func (ino *Goduino) Port() string { return ino.port }

// Name returns the  FirmataAdaptors name
func (ino *Goduino) Name() string { return ino.name }

// ServoConfig sets the pulse width in microseconds for a pin attached to a servo
func (ino *Goduino) ServoConfig(pin, min, max int) error {
	//p, err := strconv.Atoi(pin)
	//if err != nil {
	//	return err
	//}

	return ino.board.ServoConfig(pin, max, min)
}

// ServoWrite writes the 0-180 degree angle to the specified pin.
func (ino *Goduino) ServoWrite(pin int, angle byte) (err error) {
	//p, err := strconv.Atoi(pin)
	//if err != nil {
	//	return err
	//}

	if ino.board.Pins()[pin].Mode != firmata.Servo {
		err = ino.board.SetPinMode(pin, firmata.Servo)
		if err != nil {
			return err
		}
	}
	ino.logger.Printf("ServoWrite(%d, %d)\r\n", pin, int(angle))
	err = ino.board.AnalogWrite(pin, int(angle))
	return
}

// PwmWrite writes the 0-254 value to the specified pin
func (ino *Goduino) PwmWrite(pin int, level byte) (err error) {
	//p, err := strconv.Atoi(pin)
	//if err != nil {
	//	return err
	//}

	if ino.board.Pins()[pin].Mode != firmata.Pwm {
		err = ino.board.SetPinMode(pin, firmata.Pwm)
		if err != nil {
			return err
		}
	}
	ino.logger.Printf("PwmWrite(%d, %d)\r\n", pin, int(level))
	err = ino.board.AnalogWrite(pin, int(level))
	return
}


// PinMode configures the specified pin to behave either as an input or an output.
func (ino *Goduino) PinMode(pin, mode int) error {
	// Check if pin is valid
	if uint8(pin) < 0 || pin > len(ino.board.Pins()) {
		return fmt.Errorf("Invalid pin number %v\n", pin)
	}
	switch mode {
	// If mode == Input
	case Input:
		// Set pin mode
		if err := ino.board.SetPinMode(pin, mode); err != nil {
			return err
		}
		if err := ino.board.ReportDigital(pin, 1); err != nil {
			return err
		}
		<-time.After(10 * time.Millisecond)
	// If mode == Analog
	case Analog:
		pin = ino.digitalPin(pin)
		// Set pin mode
		if err := ino.board.SetPinMode(pin, mode); err != nil {
			return err
		}
		if err := ino.board.ReportAnalog(pin, 1); err != nil {
			return err
		}
		<-time.After(10 * time.Millisecond)
	default:
		// Set pin mode
		if err := ino.board.SetPinMode(pin, mode); err != nil {
			return err
		}
	}
	// PinMode was successful
	ino.logger.Printf("pinMode(%d, %s)\r\n", pin, PinMode(mode))
	return nil
}

// Close the serial connection to properly clean up after ourselves
// Usage: defer client.Close()
func (ino *Goduino) Delay(duration time.Duration) {
	time.Sleep(duration)
}

// digitalPin converts pin number to digital mapping
func (ino *Goduino) digitalPin(pin int) int {
	return pin + 14
}

type PinMode uint8

func (m PinMode) String() string {
	switch {
	case m == Input:
		return "INPUT"
	case m == Output:
		return "OUTPUT"
	case m == Analog:
		return "ANALOG"
	case m == Pwm:
		return "PWM"
	case m == Servo:
		return "SERVO"
	}
	return "UNKNOWN"
}
