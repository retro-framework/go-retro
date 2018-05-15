import './ObjectDatabaseViewer.css';

import * as React from 'react';

import ObjectDatabaseCheckpoint from './ObjectDatabaseCheckpoint';

interface IObjectDatabaseViewerProps {
  headRef: string;
}

class ObjectDatabaseViewer extends React.Component<IObjectDatabaseViewerProps, any> {
  public render() {
    if (this.props.headRef === "") {
      return (<span>Please choose a ref from above</span>);
    }
    return (<ul className="odbv">
      <ObjectDatabaseCheckpoint refHash={this.props.headRef} />
    </ul>
    );
  }
}

export default ObjectDatabaseViewer;
