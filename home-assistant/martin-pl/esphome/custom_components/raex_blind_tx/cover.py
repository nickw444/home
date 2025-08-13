import esphome.codegen as cg
import esphome.config_validation as cv
from esphome.components import cover
from esphome.const import CONF_ID
from . import raex_blind_tx_ns, RaexBlindTX, CONF_RAEX_PARENT, CONF_REMOTE_ID, CONF_CHANNEL_ID

DEPENDENCIES = ["raex_blind_tx"]

RaexBlindCover = raex_blind_tx_ns.class_("RaexBlindCover", cover.Cover, cg.Component)

CONFIG_SCHEMA = cover.cover_schema(RaexBlindCover).extend({
    # cv.GenerateID(): cv.declare_id(RaexBlindCover),
    cv.GenerateID(CONF_RAEX_PARENT): cv.use_id(RaexBlindTX),
    cv.Required(CONF_REMOTE_ID): cv.int_range(min=0, max=65535),
    cv.Required(CONF_CHANNEL_ID): cv.int_range(min=0, max=255),
}).extend(cv.COMPONENT_SCHEMA)

async def to_code(config):
    var = cg.new_Pvariable(config[CONF_ID])
    await cg.register_component(var, config)
    await cover.register_cover(var, config)

    parent = await cg.get_variable(config[CONF_RAEX_PARENT])
    cg.add(var.set_parent(parent))
    cg.add(var.set_remote_id(config[CONF_REMOTE_ID]))
    cg.add(var.set_channel_id(config[CONF_CHANNEL_ID])) 