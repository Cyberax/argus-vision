package visibility

const CanaryAttributeName = "canary"

// OTEL only defines these metrics:
// const (
//	Dimensionless string = "1"
//	Bytes         Unit = "By"
//	Milliseconds  Unit = "ms"
// )
// We add some additional units that are not defined in OTEL headers, but are supported by
// Prometheus and other collectors.
// For Prometheus see: https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/10028
//goland:noinspection GoCommentStart,GoUnusedConst
const (
	Dimensionless string = "1"

	// Time
	UnitDays         string = "d"
	UnitHours        string = "h"
	UnitMinutes      string = "min"
	UnitSeconds      string = "s"
	UnitMilliseconds string = "ms"
	UnitMicroseconds string = "us"
	UnitNanoseconds  string = "ns"

	// Bytes
	UnitKibiBytes string = "KiBy"
	UnitMebiBytes string = "MiBy"
	UnitGibiBytes string = "GiBy"
	UnitTibiBytes string = "TiBy"

	UnitBytes     string = "B"
	UnitKiloBytes string = "KB"
	UnitMegaBytes string = "MB"
	UnitGigaBytes string = "GB"
	UnitTeraBytes string = "TB"

	// Network Speed
	UnitKibiBytesSec string = "KiBy/s"
	UnitMebiBytesSec string = "MiBy/s"
	UnitGibiBytesSec string = "GiBy/s"
	UnitTibiBytesSec string = "TiBy/s"

	UnitBytesSec     string = "B/s"
	UnitKiloBytesSec string = "KB/s"
	UnitMegaBytesSec string = "MB/s"
	UnitGigaBytesSec string = "GB/s"
	UnitTeraBytesSec string = "TB/s"

	// SI
	UnitMetersPerSec string = "m/s"
	UnitMeters       string = "m"
	UnitVolts        string = "V"
	UnitAmperes      string = "A"
	UnitJoules       string = "J"
	UnitWatts        string = "W"
	UnitGrams        string = "g"

	// Misc
	UnitCelsius string = "Cel"
	UnitHertz   string = "Hz"
	UnitPercent string = "%"
	UnitDollars string = "$"
)
