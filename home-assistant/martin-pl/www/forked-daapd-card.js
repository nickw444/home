import { LitElement, html } from 'https://unpkg.com/lit-element@2.0.1/lit-element.js?module';

if (!customElements.get('ha-slider')) {
  customElements.define(
    'ha-slider',
    class extends customElements.get('paper-slider') {},
  );
}

if (!customElements.get('ha-icon-button')) {
  customElements.define(
    'ha-icon-button',
    class extends customElements.get('paper-icon-button') {},
  );
}

if (!customElements.get('ha-icon')) {
  customElements.define(
    'ha-icon',
    class extends customElements.get('iron-icon') {},
  );
}

if (!customElements.get('ha-switch')) {
  customElements.define(
    'ha-switch',
    class extends customElements.get('paper-toggle-button') {},
  );
}

class ForkedDaapdCard extends LitElement {
  constructor() {
    super();
    this._outputs = [];
    this._showOutput = true;
    this._icons = {
      'playing': {
        true: 'mdi:pause',
        false: 'mdi:play'
      },
      'prev': 'mdi:skip-previous',
      'next': 'mdi:skip-next',
      'power': 'mdi:power',
      'mute': {
        true: 'mdi:volume-off',
        false: 'mdi:volume-high'
      },
      'dropdown': {
        true: 'mdi:chevron-down',
        false: 'mdi:chevron-up',
      },
      'speaker': {
        true: 'mdi:speaker',
        false: 'mdi:speaker-off'
      }
    };
    this._media_info = [
      { attr: 'media_title' },
      { attr: 'media_artist' },
      { attr: 'media_series_title' },
      { attr: 'media_season', prefix: 'S' },
      { attr: 'media_episode', prefix: 'E'}
    ];
  }

  static get properties() {
    return {
      _hass: {},
      config: {},
      entity: {},
      _outputs: {},
      ws: {},
      _showOutput: Boolean
    };
  }

  set hass(hass) {
    const entity = hass.states[this.config.entity];
    this._hass = hass;
    if (entity && this.entity !== entity)
      this.entity = entity;
  }
  set outputs(outputs) {
    if (outputs && this._outputs !== outputs)
      this._outputs = outputs;
  }

  setConfig(config) {
    if (!config.entity || config.entity.split('.')[0] !== 'media_player')
      throw new Error('Specify an entity from within the media_player domain.');

    const cardConfig = Object.assign({
      icon: config.icon || null,
      nested: config.nested || false,
      outputs: config.outputs || null,
      ip: config.ip || '127.0.0.1',
      port: config.port || 3689,
      ws_port: config.ws_port || 3688
    }, config);

    this.config = cardConfig;
  }

  shouldUpdate(changedProps) {
    const change = changedProps.has('entity') ||
      changedProps.has('_showOutput') ||
      changedProps.has('_outputs');
    if (change) {
      if (!this.ws) this._initSocket();
      return true;
    }
  }

  async _initSocket() {
    this.ws = new WebSocket('ws://' + this.config.ip + ':' + this.config.ws_port, 'notify');
    this.ws.onopen = () => this._fetchOutputs();
    this.ws.onerror = () => this.ws = null;
    this.ws.onmessage = (message) => this._fetchOutputs();
  }

  async _fetchOutputs() {
    try {
      const resp = await fetch('http://' + this.config.ip + ':' + this.config.port + '/api/outputs');
      if (resp.ok) {
        let json = await resp.json();
        if (json.outputs) this.outputs = json.outputs;
      }
    } catch(err) {}
  }

  render({_hass, config, entity} = this) {
    if (!entity) return;
    const name = config.name || this._getAttribute('friendly_name');

    return html`
      ${this._style()}
      <ha-card ?nested=${config.nested}>
        <div id='player' class='flex'>
          <div class='flex'>
            <div class='icon'><ha-icon icon=${this._getIcon}></ha-icon></div>
            <div class='info'>
              <span class='name'>${name}</span>
              ${this._computeMediaInfo()}
            </div>
            <div class='power-state flex'>
              ${this._computePowerStrip()}
            </div>
          </div>
          <div class='flex control-row'>
            ${this._computeControls()}
          </div>
          ${this._computeOutputs()}
        </div>
      </ha-card>`;
  }

  _sortOutputs(items) {
    // return items.sort((a, b) => {
    //   return (a.name < b.name) ? -1 : (a.name > b.name) ? 1 : 0;
    // });
    return items.sort((a, b) => {
      return !a.selected && b.selected ? 1 : a.selected && !b.selected ? -1 : 0;
    });
  }

  _computeControls() {
    const vol = this.entity.attributes.volume_level * 100 || 0;
    return this._isActive ? html`
      <ha-icon-button class='dropdown'
        icon=${this._icons.dropdown[!this._showOutput]}
        @click='${(e) => { e.stopPropagation(); this._showOutput = !this._showOutput}}'>
      </ha-icon-button>
      <ha-slider class='volume-slider'
        @change='${(e) => this._handleVolumeChange(e)}'
        @click='${e => e.stopPropagation()}'
        min='0' max='100' value=${vol}
        ignore-bar-touch pin>
      </ha-slider>
      <div class='control-buttons'>
        <ha-icon-button id='prev-button' icon=${this._icons['prev']}
          @click='${(e) => this._callService(e, "media_previous_track")}'>
        </ha-icon-button>
        <ha-icon-button id='play-button'
          icon=${this._icons.playing[this._isPlaying]}
          @click='${(e) => this._callService(e, "media_play_pause")}'>
        </ha-icon-button>
        <ha-icon-button id='next-button' icon=${this._icons['next']}
          @click='${(e) => this._callService(e, "media_next_track")}'>
        </ha-icon-button>
      </div>` : '';
  }

  _computeOutputs() {
    if (!this._isActive || !this._showOutput) return;
    let outputs = this._outputs;
    if (this.config.outputs) {
      outputs = outputs.filter(output => this.config.outputs.includes(output.id));
    }

    return html`
      <div id='outputs'>
        <span class='title'>SPEAKERS</span>
        ${outputs.map(output =>
          html`
            <div class='output flex' selected=${output.selected} data-id=${output.id}>
              <div class='icon'>
                <ha-icon icon=${this._icons.speaker[output.selected]}></ha-icon>
              </div>
              <div class='info'>
                <span class='name'>${output.name}</span>
                <span class='type'>
                  ${output.type} <span class='vol'>- ${output.volume}%</span>
                </span>
              </div>
              <div class='power-state flex'>
                ${output.selected ? html`
                  <ha-slider class='volume-slider'
                    @change='${(e) => this._setOutput(e, output.id, {volume: e.target.value})}'
                    @click='${e => e.stopPropagation()}'
                    min='0' max='100' value=${output.volume}
                    ignore-bar-touch pin>
                  </ha-slider>
                ` : '' }
                <ha-switch ?checked=${output.selected}
                  @change='${(e) => this._setOutput(e, output.id, {selected: !output.selected})}'
                  @click='${e => e.stopPropagation()}'>
                </ha-switch>
              </div>
            </div>`
        )}
      </div>`;
  }

  async _setOutput(e, id, data) {
    e.stopPropagation();
    const options = {
      method: 'PUT',
      mode: 'cors',
      body: JSON.stringify(data)
    }
    await fetch('http://' + this.config.ip + ':' + this.config.port + '/api/outputs/' + id, options);
  }

  _computeMediaInfo() {
    const items = this._media_info.map(item => {
      item.info = this._getAttribute(item.attr);
      item.prefix = item.prefix || '';
      return item;
    }).filter(item => item.info !== '');

    return html`
      <div class='media-info'>
        ${items.map(item => html`<span>${item.prefix + item.info}</span>`)}
      </div>`;
  }

  _computePowerStrip({entity, config} = this) {
    if (entity.state === 'unavailable') {
      return html`
        <span id='unavailable'>
          ${this._getLabel('state.default.unavailable', 'Unavailable')}
        </span>`;
    }
    return html`${this._computePower()}`;
  }

  _computePower() {
    return html`
      <ha-icon-button id='power-button'
        icon=${this._icons['power']}
        @click='${(e) => this._callService(e, "toggle")}'
        ?color=${this._isActive}>
      </ha-icon-button>`;
  }

  _callService(e, service, options, component = 'media_player') {
    e.stopPropagation();
    options = (options === null || options === undefined) ? {} : options;
    options.entity_id = options.entity_id || this.config.entity;
    this._hass.callService(component, service, options);
  }

  _handleVolumeChange(e) {
    e.stopPropagation();
    const volPercentage = parseFloat(e.target.value);
    const vol = volPercentage > 0 ? volPercentage / 100 : 0;
    this._callService(e, 'volume_set', { volume_level: vol })
  }

  _fire(type, detail, options) {
    options = options || {};
    detail = (detail === null || detail === undefined) ? {} : detail;
    const e = new Event(type, {
      bubbles: options.bubbles === undefined ? true : options.bubbles,
      cancelable: Boolean(options.cancelable),
      composed: options.composed === undefined ? true : options.composed
    });
    e.detail = detail;
    this.dispatchEvent(e);
    return e;
  }

  get _isPaused() {
    return this.entity.state === 'paused';
  }

  get _isPlaying() {
    return this.entity.state === 'playing';
  }

  get _isActive() {
    return (this.entity.state !== 'off' && this.entity.state !== 'unavailable') || false;
  }

  get _getIcon() {
    return this.config.icon || this.entity.attributes.icon;
  }

  _hasMediaInfo() {
    const items = this._media_info.map(item => {
      return item.info = this._getAttribute(item.attr);
    }).filter(item => item !== '');
    return items.length == 0 ? false : true;
  }

  _getAttribute(attr, {entity} = this) {
    return entity.attributes[attr] || '';
  }

  _getLabel(label, fallback = 'unknown') {
    const lang = this._hass.selectedLanguage || this._hass.language;
    const resources = this._hass.resources[lang];
    return (resources && resources[label] ? resources[label] : fallback);
  }

  _style() {
    return html`
      <style>
        ha-card {
          padding: 16px;
          position: relative;
        }
        ha-card[nested] {
          background: none;
          box-shadow: none;
          padding: 0;
        }
        ha-card header {
          display: none;
        }
        #player {
          flex-flow: column;
        }
        #player ha-slider {
          min-width: 125px;
          height: 40px;
          flex: 1;
        }
        .dropdown {
          flex: 1;
          margin-right: 16px;
          width: 40px;
          flex: 0 0 40px;
        }
        .control-buttons {
          display: flex;
          flex-wrap: nowrap;
          margin-left: auto;
        }
        .flex {
          display: flex;
          display: -webkit-flex;
        }
        .justify {
          justify-content: space-between;
          -webkit-justify-content: space-between;
        }
        .icon {
          display: inline-block;
          position: relative;
          flex: 0 0 40px;
          width: 40px;
          line-height: 40px;
          text-align: center;
          color: var(--ha-item-icon-color, #44739e);
        }
        ha-icon-button[color] {
          color: var(--accent-color);
        }
        .info {
          flex: 1 0 60px;
          margin-left: 16px;
          display: flex;
          min-width: 0;
          flex-flow: column;
          justify-content: center;
        }
        .name, .media-info {
          display: inline-block;
          display: -webkit-box;
          -webkit-line-clamp: 1;
          -webkit-box-orient: vertical;
          overflow: hidden;
          word-wrap: break-word;
          word-break: break-all;
        }
        .name {
          color: var(--primary-text-color);
        }
        .media-info {
          color: var(--secondary-text-color);
        }
        .media-info span:before {
          content: ' - ';
        }
        .media-info span:first-child:before {
          content: '';
        }
        .power-state {
          margin-left: auto;
          margin-right: 0;
        }
        #unavailable {
          line-height: 40px;
        }
        .output[selected='true'] .power-state {
          justify-content: flex-end;
          flex: 2 1 60px;
        }
        .output ha-slider {
          min-width: 10px;
          max-width: 400px;
          height: 40px;
          width: auto;
          flex: 1;
          opacity: 1;
        }
        .output {
          font-size: 1rem;
          margin: 8px 0;
          opacity: 1;
          transition: opacity .25s;
        }
        .output[selected='false'] {
          opacity: .75;
        }
        .output[selected='false'] .type {
          color: var(--primary-text-color);
          opacity: .5;
        }
        .output[selected='false'] .type > span {
          display: none;
        }
        .output[selected='true'] .type {
          color: var(--accent-color);
        }
        .output ha-icon {
          width: 20px;
        }
        .type {
          opacity: 1;
          font-size: .9rem;
        }
        .title {
          display: none;
          opacity: .75;
          margin-left: 56px;
        }
        #outputs {
          padding-top: 16px;
          position: relative;
        }
        #outputs:before {
          content: '';
          position: absolute;
          left: 56px; right: 8px; top: 8px;
          height: 2px;
          background: var(--primary-background-color);
        }
        ha-toggle-button {
          cursor: pointer;
        }
      </style>
    `;
  }

  getCardSize() {
    return 1;
  }
}

customElements.define('forked-daapd-card', ForkedDaapdCard);
