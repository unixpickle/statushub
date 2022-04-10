class Settings extends React.Component {
  constructor() {
    super();
    this.state = {
      initLoading: true,
      initError: null,
      password: { loading: false, error: null },
      settings: { loading: false, error: null },
      passwordFields: {
        old: '',
        new: '',
        confirm: ''
      },
      settingsFields: {
        logSize: 0,
        mediaCache: 0
      },
    };
    this._passwordReq = null;
    this._settingsReq = null;
  }

  componentDidMount() {
    this._settingsReq = callAPI('getprefs', {}, (err, data) => {
      this._settingsReq = null;
      this.setState({
        initLoading: false,
        initError: err,
        settingsFields: data
      });
    });
  }

  componentWillUnmount() {
    if (this._settingsReq) {
      this._settingsReq.abort();
    }
    if (this._passwordReq) {
      this._passwordReq.abort();
    }
  }

  handleChangePassword() {
    this.setState({ password: { loading: true, error: null } });
    this._passwordReq = callAPI('chpass', this.state.passwordFields, (err) => {
      this._passwordReq = null;
      this.setState({ password: { loading: false, error: err, done: true } });
    });
  }

  handleSaveSettings() {
    const fields = Object.assign({}, this.state.settingsFields);
    var keys = Object.keys(fields);
    for (var i = 0; i < keys.length; ++i) {
      if ('string' === typeof fields[keys[i]]) {
        var str = fields[keys[i]];
        fields[keys[i]] = parseInt(str);
        if (isNaN(fields[keys[i]])) {
          this.setState({ settings: { loading: false, error: 'invalid number: ' + str } });
          return;
        }
      }
    }

    this.setState({ settings: { loading: true, error: null } });
    this._settingsReq = callAPI('setprefs', fields, (err) => {
      this._settingsReq = null;
      this.setState({ settings: { loading: false, error: err, done: true } });
    });
  }

  render() {
    let mainSettings;
    if (this.state.initLoading) {
      mainSettings = <Loader />;
    } else if (this.state.initError) {
      mainSettings = <label className="init-error">{this.state.initError}</label>;
    } else {
      const handleSet = (name, val) => this.handleSettingChanged(name, val);
      const handleSave = () => this.handleSaveSettings();
      mainSettings = <MainSettings data={this.state.settingsFields}
        status={this.state.settings}
        onChange={handleSet}
        onSave={handleSave} />;
    }
    return (
      <div className="settings-pane">
        <div className="password-setter">
          {this.passwordField('Old password', 'old')}
          {this.passwordField('Confirm password', 'confirm')}
          {this.passwordField('New password', 'new')}
          <SettingsAction text="Change Password" info={this.state.password}
            onAction={() => this.handleChangePassword()} />
        </div>
        <div className="main-settings">
          {mainSettings}
        </div>
      </div>
    );
  }

  passwordField(name, id) {
    const handleChange = (e) => {
      this.handlePasswordFieldChanged(id, e.target.value);
    };
    return <SettingsField name={name} type="password"
      onChange={handleChange}
      value={this.state.passwordFields[id]} />;
  }

  handlePasswordFieldChanged(id, val) {
    const f = Object.assign({}, this.state.passwordFields);
    f[id] = val;
    this.setState({ passwordFields: f });
  }

  handleSettingChanged(id, val) {
    const f = Object.assign({}, this.state.settingsFields);
    f[id] = val;
    this.setState({ settingsFields: f });
  }
}

function MainSettings(props) {
  const logSizeChanged = (e) => props.onChange('logSize', e.target.value);
  const mediaCacheChanged = (e) => props.onChange('mediaCache', e.target.value);
  return (
    <div>
      <SettingsField name="Log Size" value={props.data.logSize}
        onChange={logSizeChanged} />
      <SettingsField name="Media Cache" value={props.data.mediaCache}
        onChange={mediaCacheChanged} />
      <SettingsAction text="Save" info={props.status} onAction={props.onSave} />
    </div>
  );
}

function SettingsField(props) {
  return (
    <div className="settings-field">
      <label>{props.name}</label>
      <input type={props.type} value={props.value} onChange={props.onChange} />
    </div>
  );
}

function SettingsAction(props) {
  const info = props.info;
  let onAction = props.onAction;
  let btnClass = '';
  if (info.loading) {
    onAction = function () { };
    btnClass = 'disabled';
  }
  let statusLabel = null;
  if (info.error) {
    statusLabel = <label className="error">{info.error}</label>;
  } else if (info.done) {
    statusLabel = <label className="success">Setting saved</label>;
  }
  return (
    <div className="settings-action">
      <button onClick={onAction} className={btnClass}>{props.text}</button>
      {(info.loading ? <Loader /> : null)}
      {statusLabel}
    </div>
  );
}
