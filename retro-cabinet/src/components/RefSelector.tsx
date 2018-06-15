import * as React from 'react';
import './RefSelector.css';

class RefSelector extends React.Component<any, any> {
  constructor(props: any) {
    super(props);
    this.props = props;
  }
  public componentDidMount() {
    console.log("didmount", this.props.selectedHash);
  }
  public componentWillUpdate() {
    if(this.props.value) {
      this.props.handleChangedSelectedHeadRefHash(this.props.value);
    }
  }
  public update = (e: any) => {
    this.props.handleChangedSelectedHeadRefHash(e.target.value);
  }
  public render() {
    if (this.props.isLoading) {
      return "<span>Loading ...</span>"
    }
    const options = this.props.refs.map((ref: any) => <option key={ref.hash} value={ref.hash}>{ref.name}</option>);
    return (
      <div>
        <select className="RefSelector" value={this.props.selectedHash} onChange={this.update}>
          <option>Choose…</option>
          {options}
        </select>
        <pre>{this.props.selectedHash}</pre>
      </div>
      );
  }
}

export default RefSelector;
