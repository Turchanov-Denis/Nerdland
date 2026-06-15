package account

func (r *Repository) GetProfileByUsername(username string) (*PublicProfile, error) {
	var profile PublicProfile

	err := r.db.QueryRow(`
		SELECT
			p.account_id,
			a.email,
			p.username,
			p.display_name,
			p.bio,
			p.avatar_url,
			a.created_at
		FROM profiles p
		INNER JOIN accounts a
			ON a.id = p.account_id
		WHERE p.username = $1
	`,
		username,
	).Scan(
		&profile.AccountID,
		&profile.Email,
		&profile.Username,
		&profile.DisplayName,
		&profile.Bio,
		&profile.AvatarUrl,
		&profile.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &profile, nil
}
