import React from 'react';
import { HiPencil, HiCog, HiUserGroup, HiBell } from 'react-icons/hi';
import { FiInbox, FiSearch, FiFilter, FiCheck, FiLogOut, FiSend } from 'react-icons/fi';
import { TbBrandMessenger, TbBrandTiktok } from 'react-icons/tb';
import { BsFacebook, BsShop } from 'react-icons/bs';

export type IconKey =
  | 'inbox'
  | 'pencil'
  | 'cog'
  | 'user'
  | 'bell'
  | 'facebook'
  | 'messenger'
  | 'tiktok'
  | 'shopee'
  | 'search'
  | 'filter'
  | 'check'
  | 'logout'
  | 'send';

const iconMap: Record<IconKey, React.ComponentType<{ size?: number | string }>> = {
  inbox: FiInbox as React.ComponentType<{ size?: number | string }>,
  pencil: HiPencil as React.ComponentType<{ size?: number | string }>,
  cog: HiCog as React.ComponentType<{ size?: number | string }>,
  user: HiUserGroup as React.ComponentType<{ size?: number | string }>,
  bell: HiBell as React.ComponentType<{ size?: number | string }>,
  facebook: BsFacebook as React.ComponentType<{ size?: number | string }>,
  messenger: TbBrandMessenger as React.ComponentType<{ size?: number | string }>,
  tiktok: TbBrandTiktok as React.ComponentType<{ size?: number | string }>,
  shopee: BsShop as React.ComponentType<{ size?: number | string }>,
  search: FiSearch as React.ComponentType<{ size?: number | string }>,
  filter: FiFilter as React.ComponentType<{ size?: number | string }>,
  check: FiCheck as React.ComponentType<{ size?: number | string }>,
  logout: FiLogOut as React.ComponentType<{ size?: number | string }>,
  send: FiSend as React.ComponentType<{ size?: number | string }>,
};

export const AppIcon: React.FC<{ name: IconKey; size?: number | string }> = ({ name, size = 20 }) => {
  const IconComponent = iconMap[name];
  if (!IconComponent) return null;
  return <IconComponent size={size} />;
};
