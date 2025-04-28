import { writable } from 'svelte/store';
import type { INavigationItem } from '../interfaces/ISocialMedia';
import { 
  HomeIcon,
  SearchIcon,
  BellIcon,
  MessageSquareIcon,
  ZapIcon,
  BookmarkIcon,
  UsersIcon,
  StarIcon,
  CheckCircleIcon,
  UserIcon,
  MoreHorizontalIcon
} from 'svelte-feather-icons';

export function useNavigation() {
  // Create a store for the active navigation item
  const activeItem = writable<string>('/feed');

  // Define navigation items with proper icons
  const getNavigationItems = (): INavigationItem[] => {
    return [
      { label: 'Home', icon: 'HomeIcon', path: '/feed' },
      { label: 'Explore', icon: 'SearchIcon', path: '/explore' },
      { label: 'Notifications', icon: 'BellIcon', path: '/notifications' },
      { label: 'Messages', icon: 'MessageSquareIcon', path: '/messages' },
      { label: 'Grok', icon: 'ZapIcon', path: '/grok' },
      { label: 'Bookmarks', icon: 'BookmarkIcon', path: '/bookmarks' },
      { label: 'Communities', icon: 'UsersIcon', path: '/communities' },
      { label: 'Premium', icon: 'StarIcon', path: '/premium' },
      { label: 'Verified Orgs', icon: 'CheckCircleIcon', path: '/verified-orgs' },
      { label: 'Profile', icon: 'UserIcon', path: '/profile' },
      { label: 'More', icon: 'MoreHorizontalIcon', path: '/more' },
    ];
  };

  // Get the icon component based on its name
  const getIconComponent = (iconName: string, size = "20") => {
    switch (iconName) {
      case 'HomeIcon':
        return HomeIcon;
      case 'SearchIcon':
        return SearchIcon;
      case 'BellIcon':
        return BellIcon;
      case 'MessageSquareIcon':
        return MessageSquareIcon;
      case 'ZapIcon':
        return ZapIcon;
      case 'BookmarkIcon':
        return BookmarkIcon;
      case 'UsersIcon':
        return UsersIcon;
      case 'StarIcon':
        return StarIcon;
      case 'CheckCircleIcon':
        return CheckCircleIcon;
      case 'UserIcon':
        return UserIcon;
      case 'MoreHorizontalIcon':
        return MoreHorizontalIcon;
      default:
        return HomeIcon;
    }
  };

  // Set the active navigation item
  const setActive = (path: string) => {
    activeItem.set(path);
  };

  return {
    navigationItems: getNavigationItems(),
    activeItem,
    setActive,
    getIconComponent
  };
} 