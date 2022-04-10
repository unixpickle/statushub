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
      streamingService: null,
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
    }

    window.onpopstate = (e) => {
      this.stopStreaming();
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
        <NavBar page={this.state.page}
          onOverview={() => this.showTab('overview')}
          onMedia={() => this.showTab('mediaOverview')}
          onSettings={() => this.showTab('settings')} />
        {this.pageContent()}
      </div>
    );
  }

  pageContent() {
    switch (this.state.page) {
      case 'overview':
        return <LogScene info={this.state.overview}
          onClick={(e) => this.showServiceLog(e)} />;
      case 'mediaOverview':
        return <LogScene info={this.state.mediaOverview}
          onClick={(e) => this.showMediaLog(e)} />;
      case 'serviceLog':
        return <LogScene info={this.state.serviceLog}
          onDelete={() => this.handleDeleteService()}
          onToggleStream={() => this.handleToggleStream()}
          streaming={this.state.streamingService === this.state.serviceLogReq} />;
      case 'mediaLog':
        return <LogScene info={this.state.mediaLog}
          onClick={(info) => this.viewMediaItem(info)}
          onDelete={() => this.handleDeleteMedia()} />;
      case 'settings':
        return <Settings info={this.state.settings} />;
      case 'deleteService':
        return <DeleteService name={this.state.serviceLogReq}
          onCancel={() => this.handleDeleteServiceCancel()}
          onDone={() => this.handleDeletedService()} />;
      case 'deleteMedia':
        return <DeleteMedia name={this.state.mediaLogReq}
          onCancel={() => this.handleDeleteMediaCancel()}
          onDone={() => this.handleDeletedMedia()} />;
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
      streamingService: null,
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
  ReactDOM.render(<Root />, document.getElementById('root'));
});
