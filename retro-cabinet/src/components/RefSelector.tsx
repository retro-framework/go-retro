import * as React from 'react';
import './RefSelector.css';

import { IStoreRefSelectorState } from '../types/index';

class RefSelector extends React.Component<IStoreRefSelectorState, {}> {
  constructor(props: any) {
    super(props);
    this.props = props;
  }
  public update = (e: any) => {
    // tslint:disable-next-line:no-console
    console.log("refSelectorUpdated", e.target.value);
  }
  public render() {
    if(this.props.loading) {
      return "<span>Loading ...</span>"
    }
    const options = this.props.refs.map((ref) => <option key={ref.hash} value={ref.hash}>{ref.name}</option>);
    return (
      <select className="RefSelector" value={this.props.selectedHash} onChange={this.update}>
        {options}
      </select>
    );
  }
}

export default RefSelector;
