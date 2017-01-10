class Root extends React.Component {
  constructor() {
    super();
    this.state = {
      page: 'overview',
      overview: { error: null, entries: null },
      fullLog: { error: null, entries: null },
      serviceLog: { error: null, entries: null },
      settings: { error: null, entries: null },
      serviceLogReq: ''
    };
    if (!history.state) {
      this.replaceHistory();
    } else {
      this.state.page = history.state.page;
      this.state.serviceLogReq = history.state.serviceLogReq;
    }
    this._client = null;
  }

  componentDidMount() {
    this._client = new Client();
    this._client.onOverview = this.gotSceneData.bind(this, 'overview');
    this._client.onServiceLog = this.gotSceneData.bind(this, 'serviceLog');
    this._client.onFullLog = this.gotSceneData.bind(this, 'fullLog');
    this._client.onSettings = this.gotSceneData.bind(this, 'settings');

    window.onpopstate = e => {
      this.setState(e.state, () => this.fetchPageData());
    };

    this.fetchPageData();
  }

  componentWillUnmount() {
    this._client.close();
  }

  render() {
    return React.createElement(
      'div',
      null,
      React.createElement(NavBar, { page: this.state.page,
        onOverview: () => this.showTab('overview'),
        onFullLog: () => this.showTab('fullLog'),
        onSettings: () => this.showTab('settings') }),
      this.pageContent()
    );
  }

  pageContent() {
    switch (this.state.page) {
      case 'overview':
        return React.createElement(LogScene, { info: this.state.overview,
          onClick: e => this.showServiceLog(e) });
      case 'fullLog':
        return React.createElement(LogScene, { info: this.state.fullLog });
      case 'serviceLog':
        return React.createElement(LogScene, { info: this.state.serviceLog });
      case 'settings':
        return React.createElement(Settings, { info: this.state.settings });
    }
    throw new Error('unsupported page: ' + this.state.page);
  }

  fetchPageData() {
    if (this.state[this.state.page].entries) {
      return;
    }
    switch (this.state.page) {
      case 'overview':
        this._client.fetchOverview();
        break;
      case 'fullLog':
        this._client.fetchFullLog();
        break;
      case 'settings':
        this._client.fetchSettings();
        break;
      case 'serviceLog':
        this._client.fetchServiceLog(this.state.serviceLogReq);
        break;
    }
  }

  gotSceneData(name, err, data) {
    const obj = {};
    obj[name] = { error: err, entries: data };
    if (name === this.state.page) {
      this.setState(obj, () => this.replaceHistory());
    } else {
      this.setState(obj);
    }
  }

  showServiceLog(info) {
    this.setState({
      page: 'serviceLog',
      serviceLog: { error: null, entries: null },
      serviceLogReq: info.serviceName
    }, () => this.pushAndFetch());
  }

  showTab(name) {
    if (this.state.page == name) {
      return;
    }
    var s = { page: name };
    if (this.state[name].error) {
      s[name] = { error: null, entries: null };
    }
    this.setState(s, () => this.pushAndFetch());
  }

  pushAndFetch() {
    this.pushHistory();
    this.fetchPageData();
  }

  pushHistory() {
    history.pushState(this.historyState(), window.title, this.pageHash());
  }

  replaceHistory() {
    history.replaceState(this.historyState(), window.title, this.pageHash());
  }

  pageHash() {
    if (this.state.page === 'overview') {
      return '#';
    }
    return '#' + this.state.page;
  }

  historyState() {
    return {
      page: this.state.page,
      serviceLogReq: this.state.serviceLogReq
    };
  }
}

window.addEventListener('load', function () {
  ReactDOM.render(React.createElement(Root, null), document.getElementById('root'));
});
class Client {
  constructor() {
    this.close();
  }

  fetchOverview() {
    setTimeout(() => {
      this.onOverview(null, [{ serviceName: 'FooService', id: 0, message: 'deleting /foo/bar' }, { serviceName: 'NetService', id: 1, message: 'Current cost: 0.05' }, { serviceName: 'NetService1', id: 2, message: 'Current cost: 0.92' }]);
    }, 1000);
  }

  fetchServiceLog(name) {
    setTimeout(() => {
      this.onServiceLog(null, [{ serviceName: name, id: 4, message: 'This is a log message.' }, { serviceName: name, id: 5, message: 'The quick brown fox.' }]);
    }, 1000);
  }

  fetchSettings() {
    setTimeout(() => {
      this.onSettings(null, {
        maxLog: 1000
      });
    });
  }

  fetchFullLog() {
    setTimeout(() => {
      this.onFullLog('network failure', null);
    }, 1000);
  }

  close() {
    this.onOverview = function () {};
    this.onServiceLog = function () {};
    this.onFullLog = function () {};
    this.onSettings = function () {};
  }
}

function callAPI(name, params, cb) {
  let canceled = false;
  const req = new XMLHttpRequest();
  req.open('POST', '/api/' + name, true);
  req.setRequestHeader('Content-Type', 'application/json');
  req.onreadystatechange = () => {
    if (req.readyState === 4) {
      try {
        const obj = JSON.parse(req.responseText);
        if (obj.error) {
          cb(obj.error, null);
        } else {
          cb(null, obj.data);
        }
      } catch (e) {
        cb('invalid JSON in response', null);
      }
    }
  };
  req.send(JSON.stringify(params));
  return req;
}
function Loader(props) {
  return React.createElement(
    'div',
    { className: 'loader' },
    'Loading'
  );
}
function LogScene(props) {
  const info = props.info;
  if (info.entries) {
    return React.createElement(LogPane, { items: info.entries, onClick: props.onClick });
  } else if (info.error) {
    return React.createElement(
      'div',
      { className: 'log-scene-error' },
      React.createElement(
        'label',
        null,
        'Load failed: ',
        info.error
      )
    );
  } else {
    return React.createElement(
      'div',
      { className: 'log-scene-loading' },
      React.createElement(Loader, null)
    );
  }
}

function LogPane(props) {
  const items = props.items.map(x => {
    return React.createElement(LogItem, { info: x, key: x.id, onClick: props.onClick });
  });
  return React.createElement(
    'ul',
    { className: 'log-pane' },
    items
  );
}

function LogItem(props) {
  const inf = props.info;
  const clickHandler = () => props.onClick(inf);
  return React.createElement(
    'li',
    { className: props.onClick ? 'clickable' : '', onClick: clickHandler },
    React.createElement(
      'label',
      { className: 'service-name' },
      inf.serviceName
    ),
    React.createElement(
      'label',
      { className: 'message' },
      inf.message
    )
  );
}
function NavBar(props) {
  const page = props.page;
  return React.createElement(
    'nav',
    null,
    React.createElement(VoidLink, { onClick: props.onOverview, name: 'Overview', cur: page === 'overview' }),
    React.createElement(VoidLink, { onClick: props.onFullLog, name: 'Full Log', cur: page === 'fullLog' }),
    React.createElement(VoidLink, { onClick: props.onSettings, name: 'Settings', cur: page === 'settings' })
  );
}

function VoidLink(props) {
  if (!props.cur) {
    return React.createElement(
      'a',
      { href: 'javascript:void(0)', onClick: props.onClick },
      props.name
    );
  } else {
    return React.createElement(
      'a',
      { href: 'javascript:void(0)', className: 'cur',
        onClick: props.onClick },
      props.name
    );
  }
}
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
        logSize: 0
      }
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
    this._passwordReq = callAPI('chpass', this.state.passwordFields, err => {
      this._passwordReq = null;
      this.setState({ password: { loading: false, error: err, done: true } });
    });
  }

  handleSaveSettings() {
    const fields = Object.assign({}, this.state.settingsFields);
    if ('string' === typeof fields.logSize) {
      fields.logSize = parseInt(fields.logSize);
      if (isNaN(fields.logSize)) {
        this.setState({ settings: { loading: false, error: 'invalid log size' } });
        return;
      }
    }

    this.setState({ settings: { loading: true, error: null } });
    this._settingsReq = callAPI('setprefs', fields, err => {
      this._settingsReq = null;
      this.setState({ settings: { loading: false, error: err, done: true } });
    });
  }

  render() {
    let mainSettings;
    if (this.state.initLoading) {
      mainSettings = React.createElement(Loader, null);
    } else if (this.state.initError) {
      mainSettings = React.createElement(
        'label',
        { className: 'init-error' },
        this.state.initError
      );
    } else {
      const handleSet = (name, val) => this.handleSettingChanged(name, val);
      const handleSave = () => this.handleSaveSettings();
      mainSettings = React.createElement(MainSettings, { data: this.state.settingsFields,
        status: this.state.settings,
        onChange: handleSet,
        onSave: handleSave });
    }
    return React.createElement(
      'div',
      { className: 'settings-pane' },
      React.createElement(
        'div',
        { className: 'password-setter' },
        this.passwordField('Old password', 'old'),
        this.passwordField('Confirm password', 'confirm'),
        this.passwordField('New password', 'new'),
        React.createElement(SettingsAction, { text: 'Change Password', info: this.state.password,
          onAction: () => this.handleChangePassword() })
      ),
      React.createElement(
        'div',
        { className: 'main-settings' },
        mainSettings
      )
    );
  }

  passwordField(name, id) {
    const handleChange = e => {
      this.handlePasswordFieldChanged(id, e.target.value);
    };
    return React.createElement(SettingsField, { name: name, type: 'password',
      onChange: handleChange,
      value: this.state.passwordFields[id] });
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
  const handleChange = e => props.onChange('logSize', e.target.value);
  return React.createElement(
    'div',
    null,
    React.createElement(SettingsField, { name: 'Log Size', value: props.data.logSize,
      onChange: handleChange }),
    React.createElement(SettingsAction, { text: 'Save', info: props.status, onAction: props.onSave })
  );
}

function SettingsField(props) {
  return React.createElement(
    'div',
    { className: 'settings-field' },
    React.createElement(
      'label',
      null,
      props.name
    ),
    React.createElement('input', { type: props.type, value: props.value, onChange: props.onChange })
  );
}

function SettingsAction(props) {
  const info = props.info;
  let onAction = props.onAction;
  let btnClass = '';
  if (info.loading) {
    onAction = function () {};
    btnClass = 'disabled';
  }
  let statusLabel = null;
  if (info.error) {
    statusLabel = React.createElement(
      'label',
      { className: 'error' },
      info.error
    );
  } else if (info.done) {
    statusLabel = React.createElement(
      'label',
      { className: 'success' },
      'Setting saved'
    );
  }
  return React.createElement(
    'div',
    { className: 'settings-action' },
    React.createElement(
      'button',
      { onClick: onAction, className: btnClass },
      props.text
    ),
    info.loading ? React.createElement(Loader, null) : null,
    statusLabel
  );
}
