import { IState as ObjectDatabaseViewerState } from './ObjectDatabaseViewer';
import { IState as RefSelectorState } from './RefSelector';
import { IState as ServerURLState } from './ServerURL';

export default interface IStoreState {
  readonly server: ServerURLState,
  readonly refSelector: RefSelectorState,
  readonly odbViewer: ObjectDatabaseViewerState,
}
