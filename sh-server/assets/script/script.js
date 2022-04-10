class Root extends React.Component {
  constructor() {
    super();
    this.state = {
      page: 'overview',
      overview: { error: null, entries: null },
      mediaOverview: { error: null, entries: null },
      serviceLog: { error: null, entries: null },
      mediaLog: { error: null, entries: null },
      serviceLogReq: '',
      mediaLogReq: '',
      streamingService: null
    };
    if (!history.state) {
      this.replaceHistory();
    } else {
      this.state.page = history.state.page;
      this.state.serviceLogReq = history.state.serviceLogReq;
      this.state.mediaLogReq = history.state.mediaLogReq;
    }
    this._client = null;
  }

  componentDidMount() {
    this._client = new Client();
    this._client.onOverview = this.gotSceneData.bind(this, 'overview');
    this._client.onServiceLog = this.gotSceneData.bind(this, 'serviceLog');
    this._client.onMediaOverview = this.gotSceneData.bind(this, 'mediaOverview');
    this._client.onMediaLog = this.gotSceneData.bind(this, 'mediaLog');
    this._client.onMediaLogStream = (name, data) => {
      if (name == this.state.serviceLogReq) {
        this.gotSceneData('serviceLog', null, data);
      }
    };
    this._client.onMediaLogStreamError = (name, err) => {
      if (name == this.state.serviceLogReq) {
        this.setState({ streamingService: null });
        this.gotSceneData('serviceLog', err, null);
      }
    };

    window.onpopstate = e => {
      this.stopStreaming();
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
        onMedia: () => this.showTab('mediaOverview'),
        onSettings: () => this.showTab('settings') }),
      this.pageContent()
    );
  }

  pageContent() {
    switch (this.state.page) {
      case 'overview':
        return React.createElement(LogScene, { info: this.state.overview,
          onClick: e => this.showServiceLog(e) });
      case 'mediaOverview':
        return React.createElement(LogScene, { info: this.state.mediaOverview,
          onClick: e => this.showMediaLog(e) });
      case 'serviceLog':
        return React.createElement(LogScene, { info: this.state.serviceLog,
          onDelete: () => this.handleDeleteService(),
          onToggleStream: () => this.handleToggleStream(),
          streaming: this.state.streamingService === this.state.serviceLogReq });
      case 'mediaLog':
        return React.createElement(LogScene, { info: this.state.mediaLog,
          onClick: info => this.viewMediaItem(info),
          onDelete: () => this.handleDeleteMedia() });
      case 'settings':
        return React.createElement(Settings, { info: this.state.settings });
      case 'deleteService':
        return React.createElement(DeleteService, { name: this.state.serviceLogReq,
          onCancel: () => this.handleDeleteServiceCancel(),
          onDone: () => this.handleDeletedService() });
      case 'deleteMedia':
        return React.createElement(DeleteMedia, { name: this.state.mediaLogReq,
          onCancel: () => this.handleDeleteMediaCancel(),
          onDone: () => this.handleDeletedMedia() });
    }
    throw new Error('unsupported page: ' + this.state.page);
  }

  fetchPageData() {
    if (this.state[this.state.page] && this.state[this.state.page].entries) {
      return;
    }
    switch (this.state.page) {
      case 'overview':
        this._client.fetchOverview();
        break;
      case 'mediaOverview':
        this._client.fetchMediaOverview();
        break;
      case 'serviceLog':
        this._client.fetchServiceLog(this.state.serviceLogReq);
        break;
      case 'mediaLog':
        this._client.fetchMediaLog(this.state.mediaLogReq);
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
    this.stopStreaming();
    this.setState({
      page: 'serviceLog',
      serviceLog: { error: null, entries: null },
      serviceLogReq: info.serviceName,
      streamingService: null
    }, () => this.pushAndFetch());
  }

  showMediaLog(info) {
    this.stopStreaming();
    this.setState({
      page: 'mediaLog',
      mediaLog: { error: null, entries: null },
      mediaLogReq: info.folder
    }, () => this.pushAndFetch());
  }

  showTab(name) {
    if (this.state.page == name) {
      return;
    }
    this.stopStreaming();
    var s = { page: name };
    if (name !== 'settings') {
      s[name] = { error: null, entries: null };
    }
    this.setState(s, () => this.pushAndFetch());
  }

  handleDeleteService() {
    this.setState({ page: 'deleteService' }, () => this.pushHistory());
  }

  handleDeleteServiceCancel() {
    this.setState({ page: 'serviceLog' }, () => this.pushHistory());
  }

  handleDeletedService() {
    this.showTab('overview');
  }

  handleToggleStream() {
    if (this.state.streamingService == this.state.serviceLogReq) {
      this.stopStreaming();
    } else {
      this.startStreaming();
    }
  }

  handleDeleteMedia() {
    this.setState({ page: 'deleteMedia' }, () => this.pushHistory());
  }

  handleDeleteMediaCancel() {
    this.setState({ page: 'mediaLog' }, () => this.pushHistory());
  }

  handleDeletedMedia() {
    this.showTab('mediaOverview');
  }

  viewMediaItem(info) {
    window.open(mediaItemURL(info.id), '_blank');
  }

  stopStreaming() {
    this._client.stopStreaming();
    this.setState({ streamingService: null });
  }

  startStreaming() {
    this._client.startStreaming(this.state.serviceLogReq);
    this.setState({ streamingService: this.state.serviceLogReq });
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
      serviceLogReq: this.state.serviceLogReq,
      mediaLogReq: this.state.mediaLogReq
    };
  }
}

window.addEventListener('load', function () {
  ReactDOM.render(React.createElement(Root, null), document.getElementById('root'));
});
class Client {
  constructor() {
    this.close();
    this._stream = null;
  }

  fetchOverview() {
    callAPI('overview', {}, (e, d) => this.onOverview(e, d));
  }

  fetchServiceLog(name) {
    callAPI('serviceLog', { service: name }, (e, d) => this.onServiceLog(e, d));
  }

  fetchMediaOverview() {
    callAPI('mediaOverview', {}, (e, d) => this.onMediaOverview(e, d));
  }

  fetchMediaLog(name) {
    callAPI('mediaLog', { folder: name }, (e, d) => this.onMediaLog(e, d));
  }

  startStreaming(name) {
    if (this._stream) {
      this._stream.stop();
      this._stream = null;
    }
    this._stream = new ServiceStream(name);
    this._stream.onchange = log => this.onMediaLogStream(name, log);
    this._stream.onerror = err => {
      this._stream = null;
      this.onMediaLogStreamError(name, err);
    };
    this._stream.start();
  }

  stopStreaming() {
    if (this._stream) {
      this._stream.stop();
    }
    this._stream = null;
  }

  close() {
    this.onOverview = () => null;
    this.onServiceLog = () => null;
    this.onMediaOverview = () => null;
    this.onMediaLog = () => null;
    this.onMediaLogStream = () => null;
    this.onMediaLogStreamError = () => null;
  }
}

function callAPI(name, params, cb) {
  const req = new XMLHttpRequest();
  req.open('POST', '/api/' + name, true);
  req.setRequestHeader('Content-Type', 'application/json');
  req.onreadystatechange = () => {
    if (req.readyState === 4) {
      let obj;
      try {
        obj = JSON.parse(req.responseText);
      } catch (e) {
        cb('invalid JSON in response', null);
        return;
      }
      if (obj.error) {
        cb(obj.error, null);
      } else {
        cb(null, obj.data);
      }
    }
  };
  req.send(JSON.stringify(params));
  return req;
}

class ServiceStream {
  constructor(serviceName) {
    this.onchange = _ => null;
    this.onerror = _ => null;
    this.serviceName = serviceName;
    this._lastLog = null;
    this._socket = null;

    this._refreshWaiting = false;
    this._refreshNeeded = false;
    this._pendingEvents = [];
  }

  start() {
    const socket = new WebSocket((location.protocol == 'https' ? 'wss' : 'ws') + '://' + location.host + '/api/serviceStream?service=' + encodeURIComponent(this.serviceName));

    socket.addEventListener('open', () => {
      if (socket !== this._socket) {
        return;
      }
      this._refreshLog();
    });

    let firstMsg = true;
    socket.addEventListener('message', event => {
      if (socket !== this._socket) {
        return;
      }
      const msg = JSON.parse(event.data);
      if (firstMsg) {
        this._pendingEvents.push(msg);
        this._refreshLog();
        firstMsg = false;
      } else if (this._refreshWaiting) {
        this._pendingEvents.push(msg);
      } else {
        this._handleEvent(msg);
      }
    });

    this._socket = socket;
  }

  stop() {
    if (this._socket !== null) {
      this._socket.close();
      this._socket = null;
      this._refreshNeeded = false;
      this._pendingEvents = [];
    }
  }

  log() {
    return this._lastLog;
  }

  isRunning() {
    return this._socket !== null;
  }

  _refreshLog() {
    if (this._refreshWaiting) {
      this._refreshNeeded = true;
      return;
    }
    this._refreshWaiting = true;
    this._refreshNeeded = false;
    callAPI('serviceLog', { service: this.serviceName }, (e, d) => {
      this._refreshWaiting = false;
      if (!this.isRunning()) {
        this._refreshNeeded = false;
      } else if (e) {
        this._refreshNeeded = false;
        this.stop();
        this.onerror(e);
      } else {
        if (this._refreshNeeded) {
          // This refresh might contain stale data.
          this._refreshLog();
        } else {
          this._handleRefresh(d);
        }
        while (this._pendingEvents.length) {
          this._handleEvent(this._pendingEvents.shift());
        }
      }
    });
  }

  _handleRefresh(newLog) {
    if (this._lastLog === null) {
      this._lastLog = newLog;
      this.onchange(this._lastLog);
      return;
    }
    const idMap = {};
    this._lastLog.forEach(x => {
      idMap[x.id] = true;
    });
    if (newLog.some(x => !idMap[x.id])) {
      this._lastLog = newLog;
      this.onchange(this._lastLog);
    }
  }

  _handleEvent(data) {
    if (!this._lastLog.some(x => x.id == data.id)) {
      this._lastLog.unshift(data);
      this.onchange(this._lastLog);
    }
  }
}

function mediaItemURL(id) {
  return '/api/mediaView?id=' + id;
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
    if (info.entries.length === 0) {
      return React.createElement(
        'div',
        { className: 'log-empty' },
        'No log entries'
      );
    } else {
      return React.createElement(LogPane, { items: info.entries, onClick: props.onClick, onDelete: props.onDelete,
        streaming: props.streaming, onToggleStream: props.onToggleStream });
    }
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
  if (props.onDelete || props.onToggleStream) {
    const streaming = props.streaming;
    const streamClass = "stream-button stream-button-" + (streaming ? "active" : "paused");
    const streamText = streaming ? 'Stop Streaming' : 'Start Streaming';
    const action = React.createElement(
      'li',
      { className: 'action', key: 'deleteaction' },
      !props.onDelete ? null : React.createElement(
        'button',
        { className: 'delete-button', onClick: props.onDelete },
        'Delete'
      ),
      !props.onToggleStream ? null : React.createElement(
        'button',
        { className: streamClass, onClick: props.onToggleStream },
        streamText
      )
    );
    items.splice(0, 0, action);
  }
  return React.createElement(
    'ul',
    { className: 'log-pane' },
    items
  );
}

function LogItem(props) {
  const inf = props.info;
  let clickHandler = !props.onClick ? null : () => props.onClick(inf);
  if (inf.hasOwnProperty('serviceName')) {
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
  } else {
    if (inf.mime.startsWith('image/')) {
      return React.createElement(
        'li',
        { className: props.onClick ? 'clickable' : '', onClick: clickHandler },
        React.createElement(
          'label',
          { className: 'service-name' },
          inf.folder
        ),
        React.createElement(
          'div',
          { className: 'message' },
          React.createElement('img', { src: mediaItemURL(inf.id), alt: inf.filename })
        )
      );
    } else {
      return React.createElement(
        'li',
        { className: props.onClick ? 'clickable' : '', onClick: clickHandler },
        React.createElement(
          'label',
          { className: 'service-name' },
          inf.folder
        ),
        React.createElement(
          'label',
          { className: 'message' },
          inf.filename,
          ' ',
          React.createElement(
            'span',
            { className: 'content-type' },
            '(',
            inf.mime,
            ')'
          )
        )
      );
    }
  }
}
function NavBar(props) {
  const page = props.page;
  return React.createElement(
    'nav',
    null,
    React.createElement(VoidLink, { onClick: props.onOverview, name: 'Overview', cur: page === 'overview' }),
    React.createElement(VoidLink, { onClick: props.onMedia, name: 'Media', cur: page === 'mediaOverview' }),
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
        logSize: 0,
        mediaCache: 0
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
  const logSizeChanged = e => props.onChange('logSize', e.target.value);
  const mediaCacheChanged = e => props.onChange('mediaCache', e.target.value);
  return React.createElement(
    'div',
    null,
    React.createElement(SettingsField, { name: 'Log Size', value: props.data.logSize,
      onChange: logSizeChanged }),
    React.createElement(SettingsField, { name: 'Media Cache', value: props.data.mediaCache,
      onChange: mediaCacheChanged }),
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
class DeleteItem extends React.Component {
  constructor() {
    super();
    this.state = {
      loading: false,
      error: null
    };
    this._deleteReq = null;
  }

  componentWillUnmount() {
    if (this._deleteReq) {
      this._deleteReq.abort();
    }
  }

  handleDelete() {
    this.setState({ loading: true, error: null });
    this._deleteReq = callAPI(this.deleteAPI(), this.queryParams(), e => {
      this._deleteReq = null;
      if (e) {
        this.setState({ loading: false, error: e });
      } else {
        this.props.onDone();
      }
    });
  }

  render() {
    let btnClass = '';
    if (this.state.loading) {
      btnClass = ' disabled';
    }
    return React.createElement(
      'div',
      { className: 'delete-pane' },
      React.createElement(
        'label',
        null,
        'Delete ',
        this.props.name,
        '?'
      ),
      React.createElement(
        'div',
        { className: 'buttons' },
        React.createElement(
          'button',
          { className: 'delete-button' + btnClass,
            onClick: () => this.handleDelete() },
          'Delete'
        ),
        React.createElement(
          'button',
          { className: 'cancel-button' + btnClass,
            onClick: () => this.props.onCancel() },
          'Cancel'
        )
      ),
      !this.state.error ? null : React.createElement(
        'label',
        { className: 'error' },
        this.state.error
      )
    );
  }

  queryParams() {
    throw Error('method is abstract');
  }

  deleteAPI() {
    throw Error('method is abstract');
  }
}

class DeleteService extends DeleteItem {
  queryParams() {
    return { service: this.props.name };
  }

  deleteAPI() {
    return 'delete';
  }
}

class DeleteMedia extends DeleteItem {
  queryParams() {
    return { folder: this.props.name };
  }

  deleteAPI() {
    return 'deleteMedia';
  }
}
