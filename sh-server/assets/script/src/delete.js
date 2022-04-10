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
    this._deleteReq = callAPI(this.deleteAPI(), this.queryParams(), (e) => {
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
    return (
      <div className="delete-pane">
        <label>Delete {this.props.name}?</label>
        <div className="buttons">
          <button className={'delete-button' + btnClass}
            onClick={() => this.handleDelete()}>Delete</button>
          <button className={'cancel-button' + btnClass}
            onClick={() => this.props.onCancel()}>Cancel</button>
        </div>
        {(!this.state.error ? null
          : <label className="error">{this.state.error}</label>)}
      </div>
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
