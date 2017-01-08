class Root extends React.Component {
  constructor() {
    super();
    this.state = {
      page: 'overview',
      overview: {error: null, entries: null},
      fullLog: {error: null, entries: null},
      serviceLog: {error: null, entries: null},
      settings: {error: null, entries: null},
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

    window.onpopstate = (e) => {
      this.setState(e.state, () => this.fetchPageData());
    };

    this.fetchPageData();
  }

  componentWillUnmount() {
    this._client.close();
  }

  render() {
    return (
      <div>
        <NavBar page={this.state.page} />
        {this.pageContent()}
      </div>
    );
  }

  pageContent() {
    switch (this.state.page) {
    case 'overview':
      return <LogScene info={this.state.overview}
                    onClick={(e) => this.showServiceLog(e)} />;
    case 'fullLog':
      // TODO: this.
      break;
    case 'serviceLog':
      return <LogScene info={this.state.serviceLog} />;
    case 'settings':
      // TODO: this.
      break;
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
    obj[name] = {error: err, entries: data};
    if (name === this.state.page) {
      this.setState(obj, () => this.replaceHistory());
    } else {
      this.setState(obj);
    }
  }

  showServiceLog(info) {
    this.setState({
      page: 'serviceLog',
      serviceLog: {error: null, entries: null},
      serviceLogReq: info.serviceName
    }, () => {
      this.pushHistory();
      this.fetchPageData();
    });
  }

  pushHistory() {
    history.pushState(this.historyState(), window.title, this.pageHash());
  }

  replaceHistory() {
    history.replaceState(this.historyState(), window.title, this.pageHash());
  }

  pageHash() {
    if (this.state.page === 'overview') {
      return '';
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

window.addEventListener('load', function() {
  ReactDOM.render(<Root />, document.getElementById('root'));
});
