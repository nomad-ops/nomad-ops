import SourceIcon from '@mui/icons-material/Source';
import KeyIcon from '@mui/icons-material/Key';
import Groups3Icon from '@mui/icons-material/Groups3';
import { NavItem } from '../domain/NavItem';
import Sources from '../pages/Sources';
import Keys from '../pages/Keys';
import Teams from '../pages/Teams';

export const MainNavItems: NavItem[] = [{
  route: "sources",
  icon: <SourceIcon />,
  page: <Sources />,
  text: "Sources"
}, {
  route: "teams",
  icon: <Groups3Icon />,
  page: <Teams />,
  text: "Teams"
}, {
  route: "keys",
  icon: <KeyIcon />,
  page: <Keys />,
  text: "Keys"
}];

export const SecondaryNavItems: NavItem[] = [];