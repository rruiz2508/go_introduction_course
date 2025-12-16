select top 10 *
from Contrato c
where c.Fecha_Ingreso > '2025-11-01'
order by c.Fecha_Ingreso desc;

select c.Contrato_Numero, ca.Articulo_Id, a.Articulo_Nombre, cac.Cuota_Nro, cac.Cuota_Monto
from Contrato c
join Contrato_Articulo ca on ca.Contrato_Id = c.Contrato_Id
join Contrato_Articulo_Cuota cac on cac.Contrato_Articulo_Id = ca.Contrato_Articulo_Id
join Articulo a on a.Articulo_Id = ca.Articulo_Id
where c.Contrato_Id = 1458;