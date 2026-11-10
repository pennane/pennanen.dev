'use server'

import { revalidatePath } from 'next/cache'
import { requireAdmin } from '../../../lib/admin/require-admin'
import { logAdmin } from '../../../lib/log/admin-api'
import type { Visibility } from '../../../lib/log/types'

export async function setVisibility(itemId: number, visibility: Visibility) {
  await requireAdmin()
  await logAdmin.changeEntry(itemId, { visibility })
  revalidatePath('/admin/entries')
}
