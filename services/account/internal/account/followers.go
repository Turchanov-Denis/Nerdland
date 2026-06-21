package account

import "errors"

type FollowService struct {
	r *Repository
}

var (
	ErrSelfFollowing = errors.New("cannot follow yourself")
)

func NewFollowService(r *Repository) *FollowService { return &FollowService{r} }

func (s *FollowService) Follow(followerID int64, followingID int64) error {
	// follow_id - the one who subscribe
	// following_id - to whom he subscribe
	// A одписался на B
	// follow list для ID это список тех на кого подписан, возвращаем  WHERE following_id = $1
	// following list для ID это список на кого подписан WHERE follower_id = $1

	if followerID == followingID {
		return ErrSelfFollowing
	}
	err := s.r.follow(followerID, followingID)
	if err != nil {
		return err
	}
	return nil
}
func (s *FollowService) Unfollow(followerID int64, followingID int64) error {
	if followerID == followingID {
		return ErrSelfFollowing
	}
	err := s.r.unFollow(followerID, followingID)
	if err != nil {
		return err
	}
	return nil
}

func (s *FollowService) GetFollowers(accountID int64) ([]PublicProfile, error) {
	list, err := s.r.GetFollowerList(accountID)
	if err != nil {
		return []PublicProfile{}, err
	}
	return list, err
}

func (s *FollowService) GetFollowing(accountID int64) ([]PublicProfile, error) {
	list, err := s.r.GetFollowingList(accountID)
	if err != nil {
		return []PublicProfile{}, err
	}
	return list, err
}
