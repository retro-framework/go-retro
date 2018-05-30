import * as React from 'react';
import { DebounceInput } from 'react-debounce-input';
import './ServerURL.css';

class ServerURL extends React.Component<any, any> { // TODO: types
  constructor(props: any) { // TODO: types
    super(props);
    this.props = props;
  }
  public componentWillMount() {
    this.props.refreshRefsFromURL((new URL(window.location.href)).origin);
  }
  public update = (e: any) => { // TODO: types
    this.props.refreshRefsFromURL(new URL(e.target.value));
  }
  public render() {
    return (
      <DebounceInput className="ServerURL" debounceTimeout={500} onChange={this.update} value={this.props.url} />
    );
  }
}

export default ServerURL;
