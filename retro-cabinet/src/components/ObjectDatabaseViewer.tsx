import './ObjectDatabaseViewer.css';

import * as React from 'react';

import ObjectDatabaseCheckpoint from './ObjectDatabaseCheckpoint';

// interface IObjectDatabaseViewerProps {
//   headRef: string;
// }

class ObjectDatabaseViewer extends React.Component<{}, any> {
  public render() {
    // if (this.props.headRef === "") {
    //   return (<span>Please choose a ref from above</span>);
    // }
    return (<ul className="odbv">
      <ObjectDatabaseCheckpoint />
    </ul>
    );
  }
}

export default ObjectDatabaseViewer;
