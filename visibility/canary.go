package visibility

import (
    "context"
    "github.com/Cyberax/argus-vision/utils"
    "go.opentelemetry.io/otel/baggage"
    "strconv"
)

const CanaryBaggageKey = "canary"

func IsCanaryRequest(ctx context.Context) bool {
    bg := baggage.FromContext(ctx)
    member := bg.Member(CanaryBaggageKey)
    if member.Value() == "" {
        return false
    }
    isCanary, err := strconv.ParseBool(member.Value())
    return err == nil && isCanary
}

func MarkAsCanary(ctx context.Context, isCanary bool) context.Context {
    mem, err := baggage.NewMember(CanaryBaggageKey, strconv.FormatBool(isCanary))
    utils.PanicIfErr(err)
    bg, err := baggage.New(mem)
    utils.PanicIfErr(err)
    return baggage.ContextWithBaggage(ctx, bg)
}
