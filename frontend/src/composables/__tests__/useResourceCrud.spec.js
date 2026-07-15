import { describe, expect, it, vi } from 'vitest';

import { useResourceCrud } from '../useResourceCrud.js';

// useResourceCrud holds only plain refs (no onMounted/lifecycle), so it can
// be exercised directly without a component host, unlike useAsyncResource.
function setup(overrides = {}) {
  const createFn = vi.fn().mockResolvedValue({ ID: '1' });
  const updateFn = vi.fn().mockResolvedValue({ ID: '1' });
  const deleteFn = vi.fn().mockResolvedValue(undefined);
  const refresh = vi.fn().mockResolvedValue(undefined);

  const crud = useResourceCrud({
    createFn,
    updateFn,
    deleteFn,
    refresh,
    emptyValues: { Name: '' },
    ...overrides,
  });

  return { crud, createFn, updateFn, deleteFn, refresh };
}

describe('useResourceCrud', () => {
  it('openCreateForm sets create mode, empty values, and clears editingId', () => {
    const { crud } = setup({ emptyValues: { Name: '', Price: '' } });

    crud.formValues.value = { Name: 'stale', Price: '999' };
    crud.editingId.value = 'stale-id';

    crud.openCreateForm();

    expect(crud.formMode.value).toBe('create');
    expect(crud.formValues.value).toEqual({ Name: '', Price: '' });
    expect(crud.editingId.value).toBe(null);
    expect(crud.isFormOpen.value).toBe(true);
    expect(crud.formError.value).toBe('');
  });

  it('openEditForm sets edit mode and maps the item through toFormValues', () => {
    const toFormValues = vi.fn((item) => ({ Name: item.Name, Price: item.Price ?? '' }));
    const { crud } = setup({ toFormValues });

    crud.openEditForm({ ID: '42', Name: 'Pan casero', Price: null });

    expect(toFormValues).toHaveBeenCalledWith({ ID: '42', Name: 'Pan casero', Price: null });
    expect(crud.formMode.value).toBe('edit');
    expect(crud.formValues.value).toEqual({ Name: 'Pan casero', Price: '' });
    expect(crud.editingId.value).toBe('42');
    expect(crud.isFormOpen.value).toBe(true);
  });

  it('handleSubmit calls createFn (not updateFn) in create mode, then refreshes and closes', async () => {
    const toPayload = vi.fn((values) => ({ Name: values.Name }));
    const { crud, createFn, updateFn, refresh } = setup({ toPayload });

    crud.openCreateForm();
    await crud.handleSubmit({ Name: 'Dulce de leche' });

    expect(createFn).toHaveBeenCalledWith({ Name: 'Dulce de leche' });
    expect(updateFn).not.toHaveBeenCalled();
    expect(refresh).toHaveBeenCalledTimes(1);
    expect(crud.isFormOpen.value).toBe(false);
  });

  it('handleSubmit calls updateFn (not createFn) in edit mode with the editing id', async () => {
    const { crud, createFn, updateFn, refresh } = setup();

    crud.openEditForm({ ID: '7', Name: 'Producto' });
    await crud.handleSubmit({ Name: 'Producto editado' });

    expect(updateFn).toHaveBeenCalledWith('7', { Name: 'Producto editado' });
    expect(createFn).not.toHaveBeenCalled();
    expect(refresh).toHaveBeenCalledTimes(1);
    expect(crud.isFormOpen.value).toBe(false);
  });

  it('a failing submit sets formError from the error body and does NOT close the form or refresh', async () => {
    const createFn = vi.fn().mockRejectedValue({ body: { message: 'Nombre invalido' } });
    const { crud, refresh } = setup({ createFn });

    crud.openCreateForm();
    await crud.handleSubmit({ Name: '' });

    expect(crud.formError.value).toBe('Nombre invalido');
    expect(crud.isFormOpen.value).toBe(true);
    expect(refresh).not.toHaveBeenCalled();
  });

  it('a failing submit without an error body message falls back to the configured errorMessage', async () => {
    const createFn = vi.fn().mockRejectedValue(new Error('network down'));
    const { crud } = setup({ createFn, errorMessage: 'No se pudo guardar el producto.' });

    crud.openCreateForm();
    await crud.handleSubmit({ Name: 'x' });

    expect(crud.formError.value).toBe('No se pudo guardar el producto.');
  });

  it('confirmDelete calls deleteFn with the requested id and refreshes', async () => {
    const { crud, deleteFn, refresh } = setup();

    crud.requestDelete('9');
    expect(crud.confirmingId.value).toBe('9');

    await crud.confirmDelete();

    expect(deleteFn).toHaveBeenCalledWith('9');
    expect(refresh).toHaveBeenCalledTimes(1);
    expect(crud.confirmingId.value).toBe(null);
  });

  it('cancelDelete clears the confirming id without calling deleteFn', () => {
    const { crud, deleteFn } = setup();

    crud.requestDelete('3');
    crud.cancelDelete();

    expect(crud.confirmingId.value).toBe(null);
    expect(deleteFn).not.toHaveBeenCalled();
  });
});
