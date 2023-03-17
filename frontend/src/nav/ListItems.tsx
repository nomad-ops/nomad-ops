import SourceIcon from '@mui/icons-material/Source';
import KeyIcon from '@mui/icons-material/Key';
import { NavItem } from '../domain/NavItem';
import Sources from '../pages/Sources';
import Keys from '../pages/Keys';

export const MainNavItems: NavItem[] = [{
  route: "sources",
  icon: <SourceIcon />,
  page: <Sources />,
  text: "Sources"
}, {
  route: "keys",
  icon: <KeyIcon />,
  page: <Keys />,
  text: "Keys"
}];

export const SecondaryNavItems: NavItem[] = [];