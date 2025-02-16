package simconnect

import (
	"fmt"
	"strconv"
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	simconnect_data "github.com/karlockhart/simconnect-go/simconnect-data"
)

func TestExample(t *testing.T) {
	instance, err := NewSimConnect("data")
	require.NoError(t, err)

	report, err := instance.GetReport()
	require.NoError(t, err)

	fmt.Printf("User Altitude: %f\n", report.Altitude)
	fmt.Printf("Eng 1: %v\n", report.Engine1Combustion)
	fmt.Printf("Eng 2: %v\n", report.Engine2Combustion)
	fmt.Printf("Eng 3: %v\n", report.Engine3Combustion)
	fmt.Printf("Eng 4: %v\n", report.Engine4Combustion)
	fmt.Printf("Engine Count: %v\n", report.EngineCount)

	err = instance.Close()
	assert.NoError(t, err)
}

// These aren't 'real' tests. This is simply for testing easily within the game.
func TestWork(t *testing.T) {
	instance, _ := NewSimConnect("data")

	instance.GetReport()

	i := 10
	objID, _ := instance.LoadNonATCAircraft("Boeing 747-8i Asobo", "G-42"+strconv.FormatInt(int64(i), 10), simconnect_data.SimconnectDataInitPosition{
		Airspeed:  200,
		Altitude:  235,
		Bank:      0,
		Heading:   0,
		Latitude:  53.34974539799793,
		Longitude: -2.274003348644879,
		OnGround:  false,
		Pitch:     0,
	}, i)

	time.Sleep(10 * time.Second)

	instance.SetDataOnSimObject(*objID, []SetSimObjectDataExpose{{
		Airspeed:  200,
		Altitude:  400,
		Bank:      0,
		Heading:   0,
		Latitude:  53.34974539799793,
		Longitude: -2.274003348644879,
		OnGround:  false,
		Pitch:     10,
	}})

	data, _ := instance.GetReportOnObjectID(*objID)
	fmt.Println(data.Altitude)
	time.Sleep(10 * time.Second)
}

func TestWork2(t *testing.T) {
	instance, err := NewSimConnect("data")
	require.NoError(t, err)

	objID, err := instance.LoadParkedATCAircraft("Boeing 747-8i Asobo", "G-420", "EGCC", 100)
	require.NoError(t, err)

	time.Sleep(5 * time.Second)

	err = instance.SetAircraftFlightPlan(*objID, 1000, "C:\\Users\\Jacques\\Desktop\\EGCCLFPG")
	require.NoError(t, err)

	data, _ := instance.GetReportOnObjectID(*objID)

	time.Sleep(5 * time.Second)

	err = instance.RemoveAIObject(*objID, 10001)
	require.NoError(t, err)

	time.Sleep(60 * time.Second)

	fmt.Println(data.Altitude)
	time.Sleep(10 * time.Second)
}

func TestSystemEvent(t *testing.T) {
	instance, err := NewSimConnect("test")
	require.NoError(t, err)

	err = instance.SubscribeToSystemEvent(10, "4sec")
	assert.NoError(t, err)

	dataChan, errChan := instance.processEventData(nil)

	if data, open := <-dataChan; open {
		fmt.Println(data)
	}

	if err, open := <-errChan; open {
		fmt.Println(err)
	}

}

func TestRadioSet(t *testing.T) {
	instance, err := NewSimConnect(t.Name())
	require.NoError(t, err)

	report, err := instance.GetReport()
	require.NoError(t, err)
	require.NotNil(t, report)

	type Events struct {
		SetComRadioHz uint32
		SwapComRadio  uint32
		Plus          uint32
		Minus         uint32
	}

	events := &Events{
		SetComRadioHz: 10,
		SwapComRadio:  20,
		Plus:          60,
		Minus:         70,
	}

	report, err = instance.GetReport()
	assert.NoError(t, err)

	err = instance.MapClientEventToSimEvent(events.SetComRadioHz, "COM_STBY_RADIO_SET_HZ")
	require.NoError(t, err)

	err = instance.MapClientEventToSimEvent(events.SwapComRadio, "COM_STBY_RADIO_SWAP")
	require.NoError(t, err)

	err = instance.MapClientEventToSimEvent(events.Plus, "COM_RADIO_FRACT_INC")
	require.NoError(t, err)

	err = instance.MapClientEventToSimEvent(events.Minus, "COM_RADIO_FRACT_DEC")
	require.NoError(t, err)

	err = instance.TransmitClientID(events.Plus, 0)
	require.NoError(t, err)
	time.Sleep(2 * time.Second)

	err = instance.TransmitClientID(events.Minus, 0)
	require.NoError(t, err)
	time.Sleep(2 * time.Second)
}

func TestThrottleAxis(t *testing.T) {
	instance, err := NewSimConnect(t.Name())
	require.NoError(t, err)

	type Events struct {
		ThrottleAxis uint32
	}

	events := &Events{
		ThrottleAxis: 21,
	}

	err = instance.MapClientEventToSimEvent(events.ThrottleAxis, "AXIS_THROTTLE_SET")
	require.NoError(t, err)

	err = instance.TransmitClientID(events.ThrottleAxis, 16384)
	require.NoError(t, err)
	time.Sleep(2 * time.Second)

	err = instance.TransmitClientID(events.ThrottleAxis, -16384)
	require.NoError(t, err)
	time.Sleep(2 * time.Second)
}

func TestAPAltitude(t *testing.T) {
	instance, err := NewSimConnect(t.Name())
	require.NoError(t, err)

	type Events struct {
		APAlt uint32
	}

	events := &Events{
		APAlt: 10,
	}

	err = instance.MapClientEventToSimEvent(events.APAlt, "AP_ALT_VAR_SET_ENGLISH")
	require.NoError(t, err)

	err = instance.TransmitClientID(events.APAlt, 100)
	require.NoError(t, err)
	time.Sleep(2 * time.Second)
}

func TestAPReport(t *testing.T) {
	instance, err := NewSimConnect(t.Name())
	require.NoError(t, err)

	apReport, err := instance.GetAPReport()
	require.NoError(t, err)

	fmt.Println(apReport.APAltSlot)
}

func TestCustomReport(t *testing.T) {
	type CustomReport struct {
		simconnect_data.RecvSimobjectDataByType
		Title    [256]byte `name:"Title"`
		Altitude float64   `name:"AUTOPILOT ALTITUDE LOCK VAR" unit:"feet"`
	}

	instance, err := NewSimConnect(t.Name())
	require.NoError(t, err)

	cReport := &CustomReport{}
	err = instance.registerDataDefinition(cReport)
	require.NoError(t, err)

	definitionID, _ := instance.getDefinitionID(cReport)
	err = instance.requestDataOnSimObjectType(definitionID, definitionID, 0, simconnect_data.SIMOBJECT_TYPE_USER)
	require.NoError(t, err)

	reportData, err := instance.processSimObjectTypeData()
	require.NoError(t, err)

	ptr := reportData.(unsafe.Pointer)
	cReport = (*CustomReport)(ptr)

	fmt.Println(cReport.Altitude)
}

func TestReport(t *testing.T) {
	instance, err := NewSimConnect(t.Name())
	require.NoError(t, err)

	report, err := instance.GetReport()
	require.NoError(t, err)

	fmt.Println(report.Altitude)
}

func TestMessage(t *testing.T) {
	instance, err := NewSimConnect(t.Name() + time.Now().String())
	require.NoError(t, err)

	err = instance.SendText(1, 1, time.Now().String())
	assert.NoError(t, err)

	err = instance.Close()
	assert.NoError(t, err)
}
