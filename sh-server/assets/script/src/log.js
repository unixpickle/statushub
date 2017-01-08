function LogScene(props) {
  const info = props.info;
  if (info.entries) {
    return <LogPane items={info.entries} onClick={props.onClick} />;
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
  return <ul className="log-pane">{items}</ul>;
}

function LogItem(props) {
  const inf = props.info;
  const clickHandler = () => props.onClick(inf);
  return (
    <li className={props.onClick ? 'clickable' : ''} onClick={clickHandler}>
      <label className="service-name">{inf.serviceName}</label>
      <label className="message">{inf.message}</label>
    </li>
  );
}
