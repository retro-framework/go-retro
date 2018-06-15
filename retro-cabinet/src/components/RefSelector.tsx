import * as React from 'react';
import './RefSelector.css';

class RefSelector extends React.Component<any, any> {
  constructor(props: any) {
    super(props);
    this.props = props;
  }
  public update = (e: any) => {
    this.props.handleChangedSelectedHeadRefHash(e.target.value);
  }
  public componentDidUpdate() {
    this.props.handleChangedSelectedHeadRefHash(this.props.selectedHash);
  }
  public render() {
    if (this.props.isLoading) {
      return "<span>Loading ...</span>"
    }
    const options = this.props.refs.map((ref: any) => <option key={ref.hash.concat(ref.name)} value={ref.hash}>{ref.name}</option>);
    return (
      <select className="RefSelector" value={this.props.selectedHash} onChange={this.update}>
        <option>Choose...</option>
        {options}
      </select>
      );
  }
}

export default RefSelector;
