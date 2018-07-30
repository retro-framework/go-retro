import './ObjectDatabaseViewer.css';

import * as React from 'react';
import * as types from '../types';

import ObjectDatabaseCheckpoint from './ObjectDatabaseCheckpoint';

class ObjectDatabaseViewer extends React.Component<any, any> { // TODO: types
  /**
   *  onClick={this.props.changeSelectedCheckpoint(checkpoint)}
   */
  public changeSelectedCheckpoint = (e: any) => { // TODO: types
    this.props.changeSelectedCheckpoint(this.props.checkpoints[0]);
  }
  public render() {
    if (this.props.isLoading === true) {
      return <span>Loading…</span>
    }
    if (this.props.selectedHeadRefHash === "") {
      return (<span>Please choose a ref from above { this.props.selectedHeadRefHash }</span>);
    }
    if (!this.props.checkpoints || this.props.checkpoints.length === 0) {
      return (<span>Branch {this.props.selectedHeadRefHash} contains no checkpoints</span>);
    }
    const cps = this.props.checkpoints.map((checkpoint: types.ICheckpoint) => <ObjectDatabaseCheckpoint key={checkpoint.hash} {...checkpoint} selectCheckpointFn={this.changeSelectedCheckpoint} />);
    return (<div className="odbv">
      { cps }
    </div> 
    );
  }
}

export default ObjectDatabaseViewer;
