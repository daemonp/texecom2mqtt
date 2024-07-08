package mqtt

import (
	"fmt"

	"github.com/daemonp/texecom2mqtt/internal/types"
	"github.com/daemonp/texecom2mqtt/internal/util"
)

type Topics struct {
	prefix string
}

func NewTopics(prefix string) *Topics {
	return &Topics{prefix: prefix}
}

func (t *Topics) Status() string {
	return fmt.Sprintf("%s/status", t.prefix)
}

func (t *Topics) Config() string {
	return fmt.Sprintf("%s/config", t.prefix)
}

func (t *Topics) Area(area types.Area) string {
	return fmt.Sprintf("%s/area/%s", t.prefix, util.Slugify(area.Name))
}

func (t *Topics) AreaCommand(area types.Area) string {
	return fmt.Sprintf("%s/area/%s/command", t.prefix, util.Slugify(area.Name))
}

func (t *Topics) Zone(zone types.Zone) string {
	return fmt.Sprintf("%s/zone/%s", t.prefix, util.Slugify(zone.Name))
}

func (t *Topics) Log() string {
	return fmt.Sprintf("%s/log", t.prefix)
}

func (t *Topics) Text() string {
	return fmt.Sprintf("%s/text", t.prefix)
}

func (t *Topics) DateTime() string {
	return fmt.Sprintf("%s/datetime", t.prefix)
}

