class DeleteService extends React.Component {
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
    this.setState({loading: true, error: null});
    this._deleteReq = callAPI('delete', {service: this.props.service}, (e) => {
      this._deleteReq = null;
      if (e) {
        this.setState({loading: false, error: e});
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
        <label>Delete {this.props.service}?</label>
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
}
