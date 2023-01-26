package visibility

import "go.opentelemetry.io/otel/metric/unit"

const CanaryAttributeName = "canary"

// OTEL only defines these metrics:
// const (
//	Dimensionless unit.Unit = "1"
//	Bytes         Unit = "By"
//	Milliseconds  Unit = "ms"
// )
// We add some additional units that are not defined in OTEL headers, but are supported by
// Prometheus and other collectors.
// For Prometheus see: https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/10028
//goland:noinspection GoCommentStart,GoUnusedConst
const (
	Dimensionless unit.Unit = "1"

	// Time
	UnitDays         unit.Unit = "d"
	UnitHours        unit.Unit = "h"
	UnitMinutes      unit.Unit = "min"
	UnitSeconds      unit.Unit = "s"
	UnitMilliseconds unit.Unit = "ms"
	UnitMicroseconds unit.Unit = "us"
	UnitNanoseconds  unit.Unit = "ns"

	// Bytes
	UnitKibiBytes unit.Unit = "KiBy"
	UnitMebiBytes unit.Unit = "MiBy"
	UnitGibiBytes unit.Unit = "GiBy"
	UnitTibiBytes unit.Unit = "TiBy"

	UnitBytes     unit.Unit = "B"
	UnitKiloBytes unit.Unit = "KB"
	UnitMegaBytes unit.Unit = "MB"
	UnitGigaBytes unit.Unit = "GB"
	UnitTeraBytes unit.Unit = "TB"

	// Network Speed
	UnitKibiBytesSec unit.Unit = "KiBy/s"
	UnitMebiBytesSec unit.Unit = "MiBy/s"
	UnitGibiBytesSec unit.Unit = "GiBy/s"
	UnitTibiBytesSec unit.Unit = "TiBy/s"

	UnitBytesSec     unit.Unit = "B/s"
	UnitKiloBytesSec unit.Unit = "KB/s"
	UnitMegaBytesSec unit.Unit = "MB/s"
	UnitGigaBytesSec unit.Unit = "GB/s"
	UnitTeraBytesSec unit.Unit = "TB/s"

	// SI
	UnitMetersPerSec unit.Unit = "m/s"
	UnitMeters       unit.Unit = "m"
	UnitVolts        unit.Unit = "V"
	UnitAmperes      unit.Unit = "A"
	UnitJoules       unit.Unit = "J"
	UnitWatts        unit.Unit = "W"
	UnitGrams        unit.Unit = "g"

	// Misc
	UnitCelsius unit.Unit = "Cel"
	UnitHertz   unit.Unit = "Hz"
	UnitPercent unit.Unit = "%"
	UnitDollars unit.Unit = "$"
)
