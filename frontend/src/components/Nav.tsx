import * as React from 'react';
import { alpha, styled } from '@mui/material/styles';
import MuiDrawer from '@mui/material/Drawer';
import MuiAppBar, { AppBarProps as MuiAppBarProps } from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import List from '@mui/material/List';
import Typography from '@mui/material/Typography';
import Divider from '@mui/material/Divider';
import IconButton from '@mui/material/IconButton';
import Badge from '@mui/material/Badge';
import MenuIcon from '@mui/icons-material/Menu';
import ChevronLeftIcon from '@mui/icons-material/ChevronLeft';
import NotificationsIcon from '@mui/icons-material/Notifications';
import PersonIcon from '@mui/icons-material/Person';
import { NavItem } from '../domain/NavItem';
import ListItemButton from '@mui/material/ListItemButton';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import Button from '@mui/material/Button';
import Menu, { MenuProps } from '@mui/material/Menu';
import MenuItem from '@mui/material/MenuItem';
import EditIcon from '@mui/icons-material/Edit';
import ArchiveIcon from '@mui/icons-material/Archive';
import FileCopyIcon from '@mui/icons-material/FileCopy';
import LogoutIcon from '@mui/icons-material/Logout';
import KeyboardArrowDownIcon from '@mui/icons-material/KeyboardArrowDown';
import { useAuth } from '../services/auth/useAuth';
import { Box } from '@mui/material';


const drawerWidth: number = 240;

interface AppBarProps extends MuiAppBarProps {
    open?: boolean;
}

const StyledMenu = styled((props: MenuProps) => (
    <Menu
        elevation={0}
        anchorOrigin={{
            vertical: 'bottom',
            horizontal: 'right',
        }}
        transformOrigin={{
            vertical: 'top',
            horizontal: 'right',
        }}
        {...props}
    />
))(({ theme }) => ({
    '& .MuiPaper-root': {
        borderRadius: 6,
        marginTop: theme.spacing(1),
        minWidth: 180,
        color:
            theme.palette.mode === 'light' ? 'rgb(55, 65, 81)' : theme.palette.grey[300],
        boxShadow:
            'rgb(255, 255, 255) 0px 0px 0px 0px, rgba(0, 0, 0, 0.05) 0px 0px 0px 1px, rgba(0, 0, 0, 0.1) 0px 10px 15px -3px, rgba(0, 0, 0, 0.05) 0px 4px 6px -2px',
        '& .MuiMenu-list': {
            padding: '4px 0',
        },
        '& .MuiMenuItem-root': {
            '& .MuiSvgIcon-root': {
                fontSize: 18,
                color: theme.palette.text.secondary,
                marginRight: theme.spacing(1.5),
            },
            '&:active': {
                backgroundColor: alpha(
                    theme.palette.primary.main,
                    theme.palette.action.selectedOpacity,
                ),
            },
        },
    },
}));

const AppBar = styled(MuiAppBar, {
    shouldForwardProp: (prop) => prop !== 'open',
})<AppBarProps>(({ theme, open }) => ({
    zIndex: theme.zIndex.drawer + 1,
    transition: theme.transitions.create(['width', 'margin'], {
        easing: theme.transitions.easing.sharp,
        duration: theme.transitions.duration.leavingScreen,
    }),
    ...(open && {
        marginLeft: drawerWidth,
        width: `calc(100% - ${drawerWidth}px)`,
        transition: theme.transitions.create(['width', 'margin'], {
            easing: theme.transitions.easing.sharp,
            duration: theme.transitions.duration.enteringScreen,
        })
    }),
}));

const Drawer = styled(MuiDrawer, { shouldForwardProp: (prop) => prop !== 'open' })(
    ({ theme, open }) => ({
        '& .MuiDrawer-paper': {
            position: 'relative',
            whiteSpace: 'nowrap',
            width: drawerWidth,
            transition: theme.transitions.create('width', {
                easing: theme.transitions.easing.sharp,
                duration: theme.transitions.duration.enteringScreen,
            }),
            boxSizing: 'border-box',
            ...(!open && {
                overflowX: 'hidden',
                transition: theme.transitions.create('width', {
                    easing: theme.transitions.easing.sharp,
                    duration: theme.transitions.duration.leavingScreen,
                }),
                width: theme.spacing(7),
                [theme.breakpoints.up('sm')]: {
                    width: theme.spacing(9),
                },
            }),
        },
    }),
);

export default function Nav({ mainNavItems, secondaryNavItems }: { mainNavItems: NavItem[], secondaryNavItems: NavItem[] }) {
    const [open, setOpen] = React.useState(false);
    const toggleDrawer = () => {
        setOpen(!open);
    };


    const [accountMenuAnchorEl, setAccountMenuAnchorEl] = React.useState<null | HTMLElement>(null);
    const accountMenuOpen = Boolean(accountMenuAnchorEl);
    const handleAccountMenuClick = (event: React.MouseEvent<HTMLElement>) => {
        setAccountMenuAnchorEl(event.currentTarget);
    };
    const handleAccountMenuClose = () => {
        setAccountMenuAnchorEl(null);
    };

    const loc = useLocation();
    const activeNavClasses = "active-nav";

    let toolbarText = "Nomad-Ops";
    let filtered = mainNavItems.filter((navItem) => {
        return loc.pathname === "/" + navItem.route;
    });
    toolbarText = filtered.length > 0 ? filtered[0].text : "Nomad-Ops";

    const auth = useAuth();

    return <div>
        <AppBar position="absolute" open={open}>
            <Toolbar
                sx={{
                    pr: '24px', // keep right padding when drawer closed
                }}
            >
                <IconButton
                    edge="start"
                    color="inherit"
                    aria-label="open drawer"
                    onClick={toggleDrawer}
                    sx={{
                        marginRight: '36px',
                        ...(open && { display: 'none' }),
                    }}
                >
                    <MenuIcon />
                </IconButton>
                <Box sx={{ flexGrow: 1 }} component="div">
                    <Typography
                        component="h1"
                        variant="h6"
                        color="inherit"
                        noWrap
                    >
                        Nomad Ops
                    </Typography>
                    <Typography
                        variant="caption" display="block" gutterBottom>
                        {toolbarText}
                    </Typography>
                </Box>
                {/* <IconButton color="inherit">
                    <Badge badgeContent={4} color="secondary">
                        <NotificationsIcon />
                    </Badge>
                </IconButton> */}
                <IconButton color="inherit" onClick={handleAccountMenuClick}>
                    <Badge color="secondary">
                        <PersonIcon />
                    </Badge>
                </IconButton>
            </Toolbar>
        </AppBar>
        <Drawer variant="permanent" open={open}>
            <Toolbar
                sx={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'flex-end',
                    px: [1],
                }}
            >
                <IconButton onClick={toggleDrawer}>
                    <ChevronLeftIcon />
                </IconButton>
            </Toolbar>
            <Divider />
            <List component="nav">
                <React.Fragment>
                    {mainNavItems.map((navItem, i) => {
                        return <Link key={"mainNav-" + i} to={navItem.route} style={{
                            color: "inherit",
                            cursor: "pointer",
                            textDecoration: "none"
                        }}>
                            <ListItemButton className={loc.pathname === "/" + navItem.route ? activeNavClasses : "test"}>
                                <ListItemIcon>
                                    {navItem.icon}
                                </ListItemIcon>
                                <ListItemText primary={navItem.text} />
                            </ListItemButton>
                        </Link>
                    })}
                </React.Fragment>
                <Divider sx={{ my: 1 }} />
                <React.Fragment>
                    {secondaryNavItems.map((navItem, i) => {
                        return <ListItemButton key={"secNav-" + i}>
                            <ListItemIcon>
                                {navItem.icon}
                            </ListItemIcon>
                            <ListItemText primary={navItem.text} />
                        </ListItemButton>
                    })}
                </React.Fragment>
            </List>
        </Drawer>


        <StyledMenu
            id="demo-customized-menu"
            MenuListProps={{
                'aria-labelledby': 'demo-customized-button',
            }}
            anchorEl={accountMenuAnchorEl}
            open={accountMenuOpen}
            onClose={handleAccountMenuClose}
        >
            <MenuItem onClick={(ev) => {

            }} disableRipple>
                <PersonIcon />
                {auth.user ? auth.user.username : "No username"}
            </MenuItem>
            <Divider sx={{ my: 0.5 }} />
            <MenuItem onClick={() => {
                auth.logout().then(() => {
                    handleAccountMenuClose();
                });
            }} disableRipple>
                <LogoutIcon />
                Logout
            </MenuItem>
        </StyledMenu>
    </div>;
}