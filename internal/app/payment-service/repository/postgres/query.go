package postgres

const (
	sqlUserExists = `
		select exists(
			select 1 from users where id = $1 and is_active = true
		)
	`

	sqlGetSubscriptionTariff = `
		select
			st.id,
			st.code,
			st.title,
			st.price_money,
			st.price_vkino_coins,
			st.is_coins_payment_available,
			st.is_money_payment_available,
			st.duration_days,
			st.level
		from subscription_tariff st
		where st.id = $1
			and st.is_active = true
		limit 1
	`

	sqlListMoneyTariffs = `
		select
			st.id,
			st.code,
			st.title,
			st.price_money,
			st.price_vkino_coins,
			st.is_coins_payment_available,
			st.is_money_payment_available,
			st.duration_days,
			st.level
		from subscription_tariff st
		where st.is_active = true
			and st.is_money_payment_available = true
			and st.price_money > 0
		order by st.level, st.id
	`

	sqlCreatePayment = `
		insert into payment (
			user_id,
			product_type,
			product_ref_id,
			amount,
			currency,
			status,
			idempotency_key,
			payment_method,
			coins_spent
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		returning id, created_at, updated_at
	`

	sqlUpdatePaymentYooKassa = `
		update payment
		set
			yookassa_payment_id = $2,
			confirmation_url = $3,
			updated_at = now()
		where id = $1
	`

	sqlUpdatePaymentStatus = `
		update payment
		set
			status = $2,
			paid_at = $3,
			updated_at = now()
		where id = $1
	`

	sqlGetPaymentByID = `
		select
			id,
			user_id,
			product_type,
			product_ref_id,
			amount,
			currency,
			status,
			payment_method,
			coins_spent,
			yookassa_payment_id,
			idempotency_key,
			confirmation_url,
			paid_at,
			created_at,
			updated_at
		from payment
		where id = $1
	`

	sqlGetPaymentByYooKassaID = `
		select
			id,
			user_id,
			product_type,
			product_ref_id,
			amount,
			currency,
			status,
			payment_method,
			coins_spent,
			yookassa_payment_id,
			idempotency_key,
			confirmation_url,
			paid_at,
			created_at,
			updated_at
		from payment
		where yookassa_payment_id = $1
	`

	sqlTryRegisterWebhookEvent = `
		insert into payment_webhook_event (yookassa_payment_id, event, payload_hash)
		values ($1, $2, $3)
		on conflict (yookassa_payment_id, event) do nothing
		returning id
	`
)
