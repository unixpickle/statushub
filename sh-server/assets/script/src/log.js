function LogScene(props) {
  const info = props.info;
  if (info.entries) {
    if (info.entries.length === 0) {
      return <div className="log-empty">No log entries</div>;
    } else {
      return <LogPane items={info.entries} onClick={props.onClick}
                      onDelete={props.onDelete}/>;
    }
  } else if (info.error) {
    return (
      <div className="log-scene-error">
        <label>Load failed: {info.error}</label>
      </div>
    );
  } else {
    return (
      <div className="log-scene-loading">
        <Loader />
      </div>
    );
  }
}

function LogPane(props) {
  const items = props.items.map((x) => {
    return <LogItem info={x} key={x.id} onClick={props.onClick} />;
  });
  if (props.onDelete) {
    const action = (
      <li className="action" key="deleteaction">
        <button className="delete-button" onClick={props.onDelete}>Delete</button>
      </li>
    );
    items.splice(0, 0, action);
  }
  return <ul className="log-pane">{items}</ul>;
}

function LogItem(props) {
  const inf = props.info;
  let clickHandler = !props.onClick ? null : () => props.onClick(inf);
  return (
    <li className={props.onClick ? 'clickable' : ''} onClick={clickHandler}>
      <label className="service-name">{inf.serviceName}</label>
      <label className="message">{inf.message}</label>
    </li>
  );
}
