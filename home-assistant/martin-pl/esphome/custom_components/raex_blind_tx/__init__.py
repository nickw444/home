import esphome.codegen as cg
import esphome.config_validation as cv
from esphome.const import CONF_ID

CONF_RAEX_PARENT = "raex_parent"
CONF_REMOTE_ID = "remote_id"
CONF_CHANNEL_ID = "channel_id"
AUTO_LOAD = ["cover", "button"]
DEPENDENCIES = ["api"]

# Define namespace and component
raex_blind_tx_ns = cg.esphome_ns.namespace('raex_blind_tx')
RaexBlindTX = raex_blind_tx_ns.class_('RaexBlindTX', cg.Component)

CONFIG_SCHEMA = cv.Schema({
    cv.GenerateID(): cv.declare_id(RaexBlindTX),
}).extend(cv.COMPONENT_SCHEMA)

async def to_code(config):
    var = cg.new_Pvariable(config[CONF_ID])
    await cg.register_component(var, config)